package main

import (
	"github.com/godbus/dbus"
	"log"
)

type AboutService struct {
	bus *dbus.Conn
}

func NewAboutService(bus *dbus.Conn) *AboutService {
	bridge := new(AboutService)
	bridge.bus = bus
	return bridge
}

// Only one for now, for testing purposes only
func (a *AboutService) GetAboutData(languageTag string) (aboutData map[string]dbus.Variant, err *dbus.Error) {
	aboutData = make(map[string]dbus.Variant)
	aboutData["DeviceName"] = dbus.MakeVariant("Golang-device")
	err = nil
	return
}

func (a *AboutService) Introspect() (xml string, err *dbus.Error) {
	xml = `
		<node name="/About" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
		      xsi:noNamespaceSchemaLocation="http://www.allseenalliance.org/schemas/introspect.xsd">
		   <interface name="org.alljoyn.About">
		      <property name="Version" type="q" access="read"/>
		      <method name="GetAboutData">
		         <arg name="languageTag" type="s" direction="in"/>
		         <arg name="aboutData" type="a{sv}" direction="out"/>
		      </method>
		      <method name="GetObjectDescription">
		         <arg name="objectDescription" type="a(sas)" direction="out"/>
		      </method>
		      <signal name="Announce">
		         <arg name="version" type="q"/>
		         <arg name="port" type="q"/>
		         <arg name="objectDescription" type="a(sas)"/>
		         <arg name="metaData" type="a{sv}"/>
		      </signal>
		   </interface>
		</node>
	`
	err = nil
	return
}

func main() {
	bus, err := dbus.SystemBus()
	bus.RequestName("com.devicehive.alljoyn.test",
		dbus.NameFlagDoNotQueue)

	if err != nil {
		log.Panic(err)
	}

	aboutService := NewAboutService(bus)

	bus.Export(aboutService, "/com/devicehive/alljoyn/test/About", "org.alljoyn.About")
	bus.Export(aboutService, "/com/devicehive/alljoyn/test/About", "org.freedesktop.DBus.Introspectable")

	// Now try to register ourself in AllJoyn via dbus
	go func() {
		bridge := bus.Object("com.devicehive.alljoyn.bridge", dbus.ObjectPath("/com/devicehive/alljoyn/bridge"))
		res := bridge.Call("com.devicehive.alljoyn.bridge.AddService", 0, "/com/devicehive/alljoyn/test/About", "com.devicehive.alljoyn.test", "/About", "org.alljoyn.About")
		log.Printf("Result: %+v", res)
		res = bridge.Call("com.devicehive.alljoyn.bridge.StartAllJoyn", 0, "com.devicehive.alljoyn.test")
	}()

	select {}
}
