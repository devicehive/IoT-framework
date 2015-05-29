package main

import (
	"encoding/json"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
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
	say.Verbosef("SendNotification(name='%s',params='%s',priority=%d)\n", name, parameters, priority)
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

	var info rest.ApiInfo

	for {
		var err error
		info, err = rest.GetApiInfo(config.URL)
		if err == nil {
			say.Verbosef("API info: %+v", info)
			break
		}
		say.Infof("API info error: %s", err.Error())
		time.Sleep(5 * time.Second)
	}

	go func() {
		control := rest.NewPollAsync()
		out := make(chan rest.DeviceCmdResource, 16)

		go rest.DeviceCmdPollAsync(
			config.URL, config.DeviceID, config.AccessKey,
			info.ServerTimestamp,
			out,
			control,
		)

		for {
			select {
			case cmd := <-out:
				parameters := ""
				if cmd.Parameters != nil {
					b, err := json.Marshal(cmd.Parameters)
					if err != nil {
						say.Infof("Could not generete JSON from parameters of command %+v\nWith error %s", cmd, err.Error())
						continue
					}

					parameters = string(b)
				}
				say.Verbosef("COMMAND %s -> %s(%v)", config.URL, cmd.Command, parameters)

				bus.Emit(restObjectPath, restCommandName, uint32(cmd.Id), cmd.Command, parameters)
			}
		}
	}()

	w := NewDbusRestWrapper(config)
	bus.Export(w, "/com/devicehive/cloud", DBusConnName)

	select {}
}
