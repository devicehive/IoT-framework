package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Conn) Run(init func()) {
	for {
		log.Printf("Starting websocket loop")
		err := c.runInternal(init)
		if err != nil {
			log.Printf("Error: %s", err)
		}
		log.Printf("Stopped websocket loop")
		time.Sleep(5 * time.Second)
	}
}

func (c *Conn) runInternal(init func()) error {
	origin := "http://localhost/"
	url := c.WebSocketURL() + "/device"

	log.Printf("Connecting using WS to: %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, http.Header{"Origin": []string{origin}})

	if err != nil {
		log.Printf("Error: %s", err.Error())
		return err
	}

	c.ws = ws

	go c.writePump()
	go func() {
		for {
			m := <-c.receive
			// log.Printf("Received response: %v", m)
			go c.handleMessage(m)
		}
	}()

	go init()

	c.readPump()
	return nil
}
