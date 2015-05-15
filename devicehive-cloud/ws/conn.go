package ws

import "github.com/gorilla/websocket"
import "sync"

type ResponseHandler func(map[string]interface{})

// Conn is an middleman between the websocket connection and the hub.
type Conn struct {
	// input
	webSocketURL, deviceID string
	commandReceived        ResponseHandler

	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send          chan []byte
	receive       chan []byte
	lastCommandId int

	queue     map[int]ResponseHandler
	queueLock sync.Mutex
}

func (c *Conn) WebSocketURL() string {
	return c.webSocketURL
}

func (c *Conn) DeviceID() string {
	return c.deviceID
}

func (c *Conn) CommandReceived() ResponseHandler {
	return c.commandReceived
}
