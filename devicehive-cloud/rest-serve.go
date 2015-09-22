package main

import (
	"encoding/json"

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

	rest.DeviceNotificationInsert(w.URL, w.DeviceID, w.AccessKey, name, dat)

	return nil
}

func (w *DbusRestWrapper) UpdateCommand(id uint32, status, result string) *dbus.Error {
	rest.DeviceCmdUpdate(w.URL, w.DeviceID, w.AccessKey, id, status, result)
	return nil
}

func restImplementation(bus *dbus.Conn, config conf.Conf) {

	err := rest.DeviceRegisterEasy(config.URL, config.DeviceID, config.DeviceName)
	if err != nil {
		say.Infof("DeviceRegisterEasy error: %s", err.Error())
		return
	}

	go func() {
		nControl := rest.NewPollAsync()
		cControl := rest.NewPollAsync()
		nOut := make(chan rest.DeviceNotificationResource, 16)
		cOut := make(chan rest.DeviceCmdResource, 16)

		go rest.DeviceNotificationPollAsync(config.URL, config.DeviceID, config.AccessKey, nOut, nControl)
		go rest.DeviceCmdPollAsync(config.URL, config.DeviceID, config.AccessKey, cOut, cControl)

		for {
			select {
			case n := <-nOut:
				parameters := ""
				if n.Parameters != nil {
					b, err := json.Marshal(n.Parameters)
					if err != nil {
						say.Infof("Could not generate JSON from parameters of notification %+v\nWith error %s", n, err.Error())
						continue
					}

					parameters = string(b)
				}
				say.Verbosef("NOTIFICATION %s -> %s(%v)", config.URL, n.Notification, parameters)
				bus.Emit(restObjectPath, restCommandName, uint32(n.Id), n.Notification, parameters)
			// case c := <-cOut:
			// 	parameters := ""
			// 	if c.Parameters != nil {
			// 		b, err := json.Marshal(c.Parameters)
			// 		if err != nil {
			// 			say.Infof("Could not generate JSON from parameters of command %+v\nWith error %s", c, err.Error())
			// 			continue
			// 		}

			// 		parameters = string(b)

			// 	}
			// 	say.Verbosef("COMMAND %s -> %s(%v)", config.URL, c.Command, parameters)
			// 	bus.Emit(restObjectPath, restCommandName, uint32(c.Id), c.Command, parameters)
			// }
		}
	}()

	w := NewDbusRestWrapper(config)
	bus.Export(w, "/com/devicehive/cloud", DBusConnName)

	select {}
}
