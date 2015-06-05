package main

// #cgo CFLAGS: -Iajtcl/inc -Iajtcl/target/linux
// #cgo LDFLAGS: -Lajtcl -lajtcl
// #include <stdio.h>
// #include <aj_debug.h>
// #include <aj_guid.h>
// #include <aj_creds.h>
// #include "alljoyn.h"
import "C"
import (
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"log"
)

type AllJoynBridge struct {
	bus *dbus.Conn
	services map[string][]*introspect.Node
}

type 

func NewAllJoynBridge(bus *dbus.Conn) *AllJoynBridge {
	bridge := new(AllJoynBridge)
	bridge.bus = bus
	services = make(map[string][]*introspect.Node)

	return bridge
}

func InitAllJoyn() {
	C.AJ_Initialize()
}


func ParseArgumentOrProperty(name, access, _type string) string {
	argString := name
	if (access == "in" || access == "write") {
		argString = argString + "<"
	} else if (access == "out" || access == "read") {
		argString = argString + ">"
	} else if {
		argString = argString + "="
	}
	argString = argString + _type

	return argString
}

func ParseArguments(args []introspect.Arg) string {
	argString := ""
	for arg := range(args) {
		argString = argString + " " + ParseArgumentOrProperty(arg.Name, arg.Direction, arg.Type)
	}
	return argString
}

func ParseAllJoynInterfaces(interfaces []introspect.Interfaces) []AJ_InterfaceDescription {
	res := make([]AJ_InterfaceDescription)

	for iface := range(interfaces) {
		desc := make([]string)
		desc := append(desc, iface.Name)

		for method := range(iface.Methods) {
			methogString := "?" + method.Name
			argString := ParseArguments(method.Args)
			desc := append(desc, methogString+argString)
		}

		for signal := range(iface.Signals) {
			signalString := "!" + signal.Name
			argString := ParseArguments(signal.Args)
			desc := append(desc, signalString+argString)
		}

		for prop := range(iface.Properties) {
			propString := "@" + ParseArgumentOrProperty(prop.Name, prop.Access, prop.Type)
			desc := append(desc, propString)
		}

		
	}
}

func ParseAllJoynObject(service *introspect.Node) C.AJ_Object {
	obj := C.AJ_Object{C.CString, service.Name, ParseAllJoynInterfaces(service.Interfaces), 0, nil}
	return obj
}

func GetAllJoynObject(services []*introspect.Node) []C.AJ_Object {
	res = make([]C.AJ_Object)

	for service := range(services) {
		res := append(res, parseAllJoynObject(service))
	}

	return res
}

func (a *AllJoynBridge) StartAllJoyn(dbusService string) *dbus.Error {
	
}

func (a *AllJoynBridge) addService(service string, node introspect.Node) {
	services, ok = a.services[service]
	if ok {
		a.services[service] = append(services, node)
	} else {
		a.services[service] = []*introspect.Node { node }
	}
}

func (a *AllJoynBridge) AddService(dbusPath, dbusService, allJoynPath, allJoynInterface string) *dbus.Error {
	node, err := introspect.Call(a.bus.Object(dbusService, dbus.ObjectPath(dbusPath)))

	if err != nil {
		log.Printf("Error getting introspect from [%s, %s]: %s", dbusService, dbusPath, err)
	}

	addService(dbusService, node)

	log.Printf("Received introspect: %+v", node)

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
