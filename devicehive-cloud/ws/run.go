package ws

import (
	"net/http"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
	"github.com/gorilla/websocket"
)

func (c *Conn) Run(accessKey string, init func()) {
	for {
		c.runInternal(accessKey, init)
		time.Sleep(5 * time.Second)
	}
}

func (c *Conn) runInternal(accessKey string, init func()) error {
	origin := "http://localhost/"
	url := c.WebSocketURL() + "/device"

	say.Verbosef("Connecting using WS to %s", url)

	ws, _, err := websocket.DefaultDialer.Dial(url,
		http.Header{"Origin": []string{origin},
			"Authorization": []string{"Bearer " + accessKey}})
	if err != nil {
		say.Infof("Dial error: %s", err.Error())
		return err
	}

	c.ws = ws

	go func() {
		for m := range c.senderQ.Out() {
			say.Verbosef("THROTTLING: Message has been received from priotiorized chan: %+v", m)
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
