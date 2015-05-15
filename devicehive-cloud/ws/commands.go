package ws

import (
	"log"

	"github.com/devicehive/IoT-framework/devicehive-cloud/pqueue"
)

func (c *Conn) RegisterDevice(deviceID, deviceName string) {
	m := map[string]interface{}{
		"action":    "device/save",
		"deviceId":  deviceID,
		"deviceKey": "00000000-0000-0000-0000-000000000000",
		"device": map[string]interface{}{
			"key":    "00000000-0000-0000-0000-000000000000",
			"name":   deviceName,
			"status": "online",
			"network": map[string]interface{}{
				"name":        "default",
				"description": "default network"},
			"deviceClass": map[string]interface{}{
				"name":           "go-gateway-class",
				"version":        "0.1",
				"offlineTimeout": 10}}}

	c.SendCommand(m)
}

func (c *Conn) Authenticate() {
	m := map[string]interface{}{
		"action":    "authenticate",
		"deviceId":  c.DeviceID(),
		"deviceKey": "00000000-0000-0000-0000-000000000000"}
	c.SendCommand(m)
}

func (c *Conn) SubscribeCommands() {
	m := map[string]interface{}{
		"action": "command/subscribe"}
	c.SendCommand(m)
}

func (c *Conn) UpdateCommand(id uint32, status string, result map[string]interface{}) {
	m := map[string]interface{}{
		"action":    "command/update",
		"commandId": id,
		"command": map[string]interface{}{
			"status": status,
			"result": result,
		},
	}
	c.SendCommand(m)
}

func (c *Conn) SendNotification(name string, parameters map[string]interface{}, priority uint64) {
	m := map[string]interface{}{
		"action": "notification/insert",
		"notification": map[string]interface{}{
			"notification": name,
			"parameters":   parameters,
		},
	}

	log.Printf("\n**** SENT FROM DBUS name=%d, priority=%d, params=%+v)", name, priority, parameters)
	removed := c.senderQ.Send(pqueue.Message(m), priority)

	for _, qi := range removed {
		log.Printf("!!!! LAST SENT FROM DBUS => REMOVING FROM QUEUE, timestamp=%d, priority=%d, parameters=%+v\n", qi.Timestamp, qi.Priority, qi.Msg)
	}
}

//TODO: delete when priority queue will be tested
// func (c *Conn) oldSendNotification(name string, parameters map[string]interface{}) {
// 	m := map[string]interface{}{
// 		"action": "notification/insert",
// 		"notification": map[string]interface{}{
// 			"notification": name,
// 			"parameters":   parameters,
// 		},
// 	}
// 	c.SendCommand(m)
// }
