package ws

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/pqueue"
	"github.com/gorilla/websocket"
	"sync"
)

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

	//Priority Queue
	senderQ *pqueue.PriorityQueue
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

func New(webSocketURL, deviceID string, sendQCapacity uint64, resp ResponseHandler) (conn Conn) {

	conn.webSocketURL = webSocketURL
	conn.deviceID = deviceID
	conn.commandReceived = resp

	conn.send = make(chan []byte, maxMessageSize)
	conn.receive = make(chan []byte, maxMessageSize)
	conn.queue = make(map[int]ResponseHandler)

	pq, err := pqueue.NewPriorityQueue(sendQCapacity, make(chan pqueue.Message))
	if err != nil {
		panic(err)
	}
	conn.senderQ = pq

	return
}
