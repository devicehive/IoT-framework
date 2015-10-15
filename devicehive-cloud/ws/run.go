package ws

import (
	"net/http"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
	"github.com/gorilla/websocket"
)

func (c *Conn) Run(init func()) {
	for {
		c.runInternal(init)
		time.Sleep(5 * time.Second)
	}
}

func (c *Conn) runInternal(init func()) error {
	origin := "http://localhost/"
	url := c.WebSocketURL() + "/device"

	say.Debugf("Connecting using WS to %s", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Origin": []string{origin}})
	if err != nil {
		say.Infof("Dial error: %s", err.Error())
		return err
	}

	c.ws = ws

	go func() {
		for m := range c.senderQ.Out() {
			say.Debugf("THROTTLING: Message has been received from priotiorized chan: %+v", m)
			c.SendCommand(m)
		}
	}()

	go c.writePump()
	go func() {
		for {
			m := <-c.receive
			go c.handleMessage(m)
		}
	}()

	go init()

	c.readPump()
	return nil
}
