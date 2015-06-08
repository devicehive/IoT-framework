package main

import (
	"encoding/xml"
	"github.com/godbus/dbus/introspect"
	"strings"
)

func DummyIntrospectProvider(dbusService, dbusPath string) (*introspect.Node, error) {
	var node introspect.Node

	xmlData := `
	<node name="/sample">
	<interface name="org.alljoyn.Bus.sample">
	  <method name="Dummy">
	    <arg name="foo" type="i" direction="in"/>
	  </method>
	  <method name="cat">
	    <arg name="inStr1" type="s" direction="in"/>
	    <arg name="inStr2" type="s" direction="in"/>
	    <arg name="outStr" type="s" direction="out"/>
	  </method>
	</interface>
	</node>
		`

	err := xml.NewDecoder(strings.NewReader(xmlData)).Decode(&node)

	if err != nil {
		return nil, err
	}

	return &node, nil
}

func main() {
	allJoynBridge := NewAllJoynBridge(nil, DummyIntrospectProvider)
	allJoynBridge.AddService("/com/devicehive/alljoyn/test", "org.alljoyn.Bus.sample", "/sample", "org.alljoyn.Bus.sample")
	allJoynBridge.StartAllJoyn("org.alljoyn.Bus.sample")
	select {}
}
