package main

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"./conf"
	"./rest"
	"./ws"
	
	"github.com/godbus/dbus"
)

type DbusObjectWrapper struct {
	c *ws.Conn
}

func NewDbusObjectWrapper(c *ws.Conn) *DbusObjectWrapper {
	w := new(DbusObjectWrapper)
	w.c = c

	return w
}

func (w *DbusObjectWrapper) SendNotification(name, parameters string) {
	var dat map[string]interface{}
	b := []byte(parameters)
	b = bytes.Trim(b, "\x00")

	err := json.Unmarshal(b, &dat)

	if err != nil {
		log.Panic(err)
	}

	w.c.SendNotification(name, dat)
}

const (
	DBusConnName = "com.devicehive.cloud"
)

func main() {
	bus, err := dbus.SystemBus()
	if err != nil {
		log.Printf("Cannot get system bus with error: %s\n", err.Error())
		log.Printf("Trying to use session bus for testing purposes...\n")
		if bus, err = dbus.SessionBus(); err != nil {
			log.Fatalf("Cannot get session bus with error: %s\n", err.Error())
			return
		}
	}

	reply, err := bus.RequestName(DBusConnName, dbus.NameFlagDoNotQueue)
	switch {
	case err != nil:
		log.Fatalf("Cannot request name '%s' with error: %s\n", DBusConnName, err.Error())
	case reply != dbus.RequestNameReplyPrimaryOwner:
		log.Fatalf("The name '%s' already taken.", DBusConnName)
	}

	configFile, config, err := conf.FromArgs()
	println(configFile)
	switch {
	case err != nil:
		log.Fatalf("Cannot read configuration in `%s` with error: %s\n", configFile, err.Error())
	case configFile == "":
		log.Printf("You should specify configuration file.\n Starting with test configuration: %+v\n", config)
	default:
		log.Printf("Starting DeviceHive gateway with configuration in '%s': %+v\n", configFile, config)
	}

	var conn *ws.Conn
	for {
		info, err := rest.GetDHServerInfo(config.URL)
		if err == nil {
			c := ws.New(info.WebSocketServerUrl, config.DeviceID, func(m map[string]interface{}) {
				log.Printf("Successfully received command: %s", m)
				p := m["parameters"]
				params := ""

				if p != nil {
					b, err := json.Marshal(p)
					if err != nil {
						log.Panic(err)
					}

					params = string(b)
				}
				log.Printf("Parameters: %v", params)
				bus.Emit("/com/devicehive/cloud", "com.devicehive.cloud.CommandReceived", m["command"].(string), params)
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

	select {}
}
