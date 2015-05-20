package main

import (
	"log"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/godbus/dbus"
)

const (
	restObjectPath  = "/com/devicehive/cloud"
	restCommandName = "com.devicehive.cloud.CommandReceived"
)

type DbusRestWrapper struct {
	URL        string
	AccessKey  string
	DeviceID   string
	DeviceName string
}

func NewDbusRestWrapper(c conf.Conf) *DbusRestWrapper {
	w := new(DbusRestWrapper)
	w.URL = c.URL
	w.AccessKey = c.AccessKey
	w.DeviceID = c.DeviceID
	w.DeviceName = c.DeviceName
	return w
}

func (w *DbusRestWrapper) SendNotification(name, parameters string, priority uint64) *dbus.Error {
	log.Printf("SendNotification(name='%s',params='%s',priority=%d)\n", name, parameters, priority)
	dat, err := parseJSON(parameters)

	if err != nil {
		return newDHError(err.Error())
	}

	m := map[string]interface{}{
		"action": "notification/insert",
		"notification": map[string]interface{}{
			"notification": name,
			"parameters":   dat,
		},
	}

	rest.DeviceCmdInsert(w.URL, w.DeviceID, w.AccessKey, "SendCommand", m)

	return nil
}

func (w *DbusRestWrapper) UpdateCommand(id uint32, status string, result string) *dbus.Error {
	dat, err := parseJSON(result)

	if err != nil {
		return newDHError(err.Error())
	}

	m := map[string]interface{}{
		"action":    "command/update",
		"commandId": id,
		"command": map[string]interface{}{
			"status": status,
			"result": dat,
		},
	}

	rest.DeviceCmdInsert(w.URL, w.DeviceID, w.AccessKey, "SendCommand", m)

	return nil
}

func restImplementation(bus *dbus.Conn, config conf.Conf) {

	// listener from cloud & sender to bus
	go func() {
		control := rest.NewPollAsync()
		out := make(chan rest.DeviceCmdResource, 16)

		go rest.DeviceCmdPollAsync(config.URL, config.DeviceID, config.AccessKey, out, control)

		for {
			select {
			case cmd := <-out:
				log.Printf("CMD RECEIVED FROM CLOUD: %+v", cmd)
				bus.Emit(restObjectPath, restCommandName, uint32(cmd.Id), cmd.Command, cmd.Parameters)
			}
		}
	}()

	// listener from bus & sender to cloud
	w := NewDbusRestWrapper(config)
	bus.Export(w, "/com/devicehive/cloud", DBusConnName)
}
