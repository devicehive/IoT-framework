package ws

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/pqueue"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

func (c *Conn) RegisterDevice(deviceID, deviceName, deviceKey, networkName, networkKey, networkDesc string) {

	// device class
	dc := map[string]interface{}{
		"name":           "go-gateway-class",
		"version":        "0.1",
		"offlineTimeout": 10}

	// network (optional)
	n := map[string]interface{}{
		"name":        networkName,
		"key":         networkKey,
		"description": networkDesc}

	// device
	d := map[string]interface{}{
		"key":    deviceKey,
		"name":   deviceName,
		"status": "Online",
		"deviceClass": dc}

	// omit "network" if all fields are empty
	if len(networkName)!=0 || len(networkKey)!=0 || len(networkDesc)!=0 {
		d["network"] = n
	}

	// action message
	m := map[string]interface{}{
		"action":    "device/save",
		"deviceId":  deviceID,
		"deviceKey": deviceKey, // authentication
		"device": d}

	c.SendCommand(m)
}

func (c *Conn) Authenticate(deviceID, deviceKey string) {
	m := map[string]interface{}{
		"action":    "authenticate",
		"deviceId":  deviceID,
		"deviceKey": deviceKey}
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

	say.Debugf("\n   SENT FROM DBUS name=%d, priority=%d, params=%+v)", name, priority, parameters)
	removed := c.senderQ.Send(pqueue.Message(m), priority)

	say.If(say.VERBOSE, func() {
		say.Alwaysf("VERBOSE:THROTTLING: %s^%d(%+v)", name, priority, parameters)
		for _, qi := range removed {
			say.Alwaysf("   => REMOVED: %d ^%d(%+v)", qi.Timestamp, qi.Priority, qi.Msg)
		}
	})

}
