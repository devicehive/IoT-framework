package ws

func New(webSocketURL, deviceID string, resp ResponseHandler) (conn Conn) {

	conn.webSocketURL = webSocketURL
	conn.deviceID = deviceID
	conn.commandReceived = resp

	conn.send = make(chan []byte, maxMessageSize)
	conn.receive = make(chan []byte, maxMessageSize)
	conn.queue = make(map[int]ResponseHandler)

	return
}
