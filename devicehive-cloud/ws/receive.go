package ws

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
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
			say.Verbosef("readPump: could not receive c.ws.NextReader() with error: %s", err.Error())
			return err
		}

		buf := make([]byte, maxMessageSize)
		_, err = r.Read(buf)

		if err != nil {
			say.Verbosef("readPump: could not read c.ws.NextReader() to the buffer (%d bytes) with error: %s", maxMessageSize, err.Error())
			return err
		}

		c.receive <- buf
	}
}

func (c *Conn) handleMessage(m []byte) {
	c.queueLock.Lock()
	defer c.queueLock.Unlock()

	var dat map[string]interface{}
	m = bytes.Trim(m, "\x00")
	err := json.Unmarshal(m, &dat)
	if err != nil {
		say.Verbosef("handleMessage: could not parse JSON in (%s) with error %s", string(m), err.Error())
	}

	a := dat["action"]

	requestId := dat["requestId"]
	var r int

	if requestId != nil {
		r = int(requestId.(float64))
	}

	callBack, ok := c.queue[r]
	if !ok && (a != "command/insert") {
		say.Verbosef("handleMessage: unhandled request with id=%d", r)
		return
	}

	switch a {
	case "device/save":
		go callBack(dat)
	case "notification/insert":
		go callBack(dat)
	case "command/subscribe":
		go callBack(dat)
	case "authenticate":
		go callBack(dat)
	case "command/update":
		go callBack(dat)
	case "command/insert":
		command := dat["command"]
		go c.CommandReceived()(command.(map[string]interface{}))
	default:
		say.Verbosef("handleMessage: Unknown notification: %s", a)
	}
	delete(c.queue, r)
}
