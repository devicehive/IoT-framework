package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const CommandResponseTimeout = 5

func (c *Conn) SendCommand(command map[string]interface{}) {
	c.lastCommandId++
	requestId := c.lastCommandId
	command["requestId"] = requestId

	r := make(chan bool, 1)
	c.queueLock.Lock()
	c.queue[requestId] = func(res map[string]interface{}) {
		r <- true
	}
	c.queueLock.Unlock()

	c.postCommand(command)

	select {
	case <-time.After(CommandResponseTimeout * time.Second):
		{
			c.queueLock.Lock()
			delete(c.queue, requestId)
			c.queueLock.Unlock()

			log.Printf("Timed out waiting for response to command: %+v", command)
		}
	case <-r:
	}
}

func (c *Conn) postCommand(command map[string]interface{}) {
	b, err := json.Marshal(command)
	if err != nil {
		log.Panic(err)
	}
	// log.Printf("Post command: %+v", command)
	c.send <- b
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Conn) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			// log.Printf("writePump(): sending: %v", message)
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				log.Print(err)
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				log.Print(err)
				return
			}
		}
	}
}

// write writes a message with the given message type and payload.
func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}
