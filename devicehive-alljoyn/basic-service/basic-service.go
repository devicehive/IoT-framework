package main

import (
	"github.com/godbus/dbus"
	"log"
)

type BasicService struct {
	bus *dbus.Conn
}

func NewBasicService(bus *dbus.Conn) *BasicService {
	bridge := new(BasicService)
	bridge.bus = bus
	return bridge
}

func (a *BasicService) GetAboutData(languageTag string) (aboutData map[string]dbus.Variant, err *dbus.Error) {
	data := make(map[string]dbus.Variant)
	data["DeviceName"] = dbus.MakeVariant("Golang-device")
	return data, nil
}

func (a *BasicService) Cat(inStr1, inStr2 string) (res string, err *dbus.Error) {
	return inStr1 + " Dear " + inStr2, nil
}

func (a *BasicService) Introspect() (xml string, err *dbus.Error) {
	xml = `
	<node name="/sample">
	<interface name="org.alljoyn.Bus.sample">
	  <method name="Dummy">
	    <arg name="foo" type="i" direction="in"/>
	  </method>
	  <method name="Cat">
	    <arg name="inStr1" type="s" direction="in"/>
	    <arg name="inStr2" type="s" direction="in"/>
	    <arg name="outStr" type="s" direction="out"/>
	  </method>
	</interface>
	</node>
		`
	err = nil
	return
}

func main() {
	bus, err := dbus.SystemBus()
	bus.RequestName("com.devicehive.alljoyn.test.basic",
		dbus.NameFlagDoNotQueue)

	if err != nil {
		log.Panic(err)
	}

	basicService := NewBasicService(bus)

	bus.Export(basicService, "/com/devicehive/alljoyn/test/basic", "org.alljoyn.About")
	bus.Export(basicService, "/com/devicehive/alljoyn/test/basic", "org.alljoyn.Bus.sample")
	bus.Export(basicService, "/com/devicehive/alljoyn/test/basic", "org.freedesktop.DBus.Introspectable")

	// Now try to register ourself in AllJoyn via dbus
	go func() {
		bridge := bus.Object("com.devicehive.alljoyn.bridge", dbus.ObjectPath("/com/devicehive/alljoyn/bridge"))
		res := bridge.Call("com.devicehive.alljoyn.bridge.AddService", 0, "/com/devicehive/alljoyn/test/basic", "com.devicehive.alljoyn.test.basic", "/sample", "org.alljoyn.Bus.sample")
		log.Printf("Result: %+v", res)
		res = bridge.Call("com.devicehive.alljoyn.bridge.StartAllJoyn", 0, "com.devicehive.alljoyn.test.basic")
	}()

	select {}
}
