package main

import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
)

type AllJoynBridge struct {
	bus *dbus.Conn
}

func NewAllJoynBridge(bus *dbus.Conn) *AllJoynBridge {
	bridge := new(AllJoynBridge)
	bridge.bus = bus
	return bridge
}

func (a *AllJoynBridge) RegisterService(dbusPath, dbusService, allJoynPath, allJoynInterface string) *dbus.Error {
	log.Printf("RegisterService")
	go func() {
		node, err := introspect.Call(a.bus.Object(dbusService, dbus.ObjectPath(dbusPath)))

		if err != nil {
			log.Printf("Error getting introspect from [%s, %s]: %s", dbusService, dbusPath, err)
		}

		log.Print(node)
	}()

	return nil
}

func main() {
	bus, err := dbus.SystemBus()
	bus.RequestName("com.devicehive.alljoyn",
		dbus.NameFlagDoNotQueue)

	if err != nil {
		log.Panic(err)
	}

	allJoynBridge := NewAllJoynBridge(bus)

	bus.Export(allJoynBridge, "/com/devicehive/alljoyn", "com.devicehive.alljoyn")

	select {}
}
