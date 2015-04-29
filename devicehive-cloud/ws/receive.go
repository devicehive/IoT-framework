package ws

import (
	"bytes"
	"encoding/json"
	"log"
	"time"
)

// readPump pumps messages from the websocket connection to the hub.
func (c *Conn) readPump() error {
	defer func() {
		c.ws.Close()
	}()

	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, r, err := c.ws.NextReader()
		if err != nil {
			log.Print(err)
			return err
		}

		buf := make([]byte, maxMessageSize)
		_, err = r.Read(buf)

		if err != nil {
			log.Print(err)
			return err
		}

		c.receive <- buf
	}
}

func (c *Conn) handleMessage(m []byte) {
	var dat map[string]interface{}
	m = bytes.Trim(m, "\x00")
	// log.Printf("handleMessage(): %s", m)
	err := json.Unmarshal(m, &dat)
	if err != nil {
		log.Panic(err)
	}

	a := dat["action"]

	requestId := dat["requestId"]
	var r int

	if requestId != nil {
		r = int(requestId.(float64))
	}

	switch a {
	case "device/save":
		c.queue[r](dat)
	case "notification/insert":
		c.queue[r](dat)
	case "command/subscribe":
		c.queue[r](dat)
	case "authenticate":
		c.queue[r](dat)
	case "command/update":
		c.queue[r](dat)
	case "command/insert":
		// log.Printf("Command/insert")
		command := dat["command"]
		go c.CommandReceived()(command.(map[string]interface{}))
	default:
		log.Printf("Unknown notification: %s", a)
	}
	delete(c.queue, r)
}
