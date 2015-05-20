package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/ws"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"
)

type DbusObjectWrapper struct {
	c *ws.Conn
}

func NewDbusObjectWrapper(c *ws.Conn) *DbusObjectWrapper {
	w := new(DbusObjectWrapper)
	w.c = c

	return w
}

func (w *DbusObjectWrapper) SendNotification(name, parameters string, priority uint64) *dbus.Error {
	log.Printf("SendNotification(name='%s',params='%s',priority=%d)\n", name, parameters, priority)
	dat, err := parseJSON(parameters)

	if err != nil {
		return newDHError(err.Error())
	}

	w.c.SendNotification(name, dat, priority)
	return nil
}

func (w *DbusObjectWrapper) UpdateCommand(id uint32, status string, result string) *dbus.Error {
	dat, err := parseJSON(result)

	if err != nil {
		return newDHError(err.Error())
	}

	w.c.UpdateCommand(id, status, dat)
	return nil
}

func wsImplementation(bus *dbus.Conn, config conf.Conf) {

	var conn *ws.Conn
	for {
		info, err := rest.GetApiInfo(config.URL)
		if err == nil {
			c := ws.New(info.WebSocketServerUrl, config.DeviceID, config.SendNotificatonQueueCapacity, func(m map[string]interface{}) {
				log.Printf("|| CLOUD received Command:%v\n", m)

				p := m["parameters"]
				params := ""

				if p != nil {
					b, err := json.Marshal(p)
					if err != nil {
						log.Panic(err)
					}

					params = string(b)
				}
				log.Printf("Command :%s", m["command"].(string))
				log.Printf("Parameters: %v", params)
				bus.Emit("/com/devicehive/cloud",
					"com.devicehive.cloud.CommandReceived",
					uint32(m["id"].(float64)),
					m["command"].(string),
					params)
			})
			conn = &c

			if err == nil {
				break
			}
		}
		log.Print(err)
		time.Sleep(5 * time.Second)
	}

	w := NewDbusObjectWrapper(conn)
	go conn.Run(func() {
		conn.RegisterDevice(config.DeviceID, config.DeviceName)
		conn.Authenticate()
		conn.SubscribeCommands()
	})

	bus.Export(w, "/com/devicehive/cloud", DBusConnName)

	// Introspectable
	n := &introspect.Node{
		Name: "/com/devicehive/cloud",
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:    "com.devicehive.cloud",
				Methods: introspect.Methods(w),
			},
		},
	}

	bus.Export(introspect.NewIntrospectable(n), "/com/devicehive/cloud", "org.freedesktop.DBus.Introspectable")

	select {}

}
