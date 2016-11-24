/*
  DeviceHive IoT-Framework business logic

  Copyright (C) 2016 DataArt

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at
 
      http://www.apache.org/licenses/LICENSE-2.0
 
  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/ 

package main

import "testing"
import "encoding/xml"
import "github.com/godbus/dbus/introspect"
import "strings"

func DummyIntrospectProvider(dbusService, dbusPath string) (*introspect.Node, error) {
	var node introspect.Node
	// xmlData := `
	// 	<node name="/About" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	// 	      xsi:noNamespaceSchemaLocation="http://www.allseenalliance.org/schemas/introspect.xsd">
	// 	   <interface name="org.alljoyn.About">
	// 	      <property name="Version" type="q" access="read"/>
	// 	      <method name="GetAboutData">
	// 	         <arg name="languageTag" type="s" direction="in"/>
	// 	         <arg name="aboutData" type="a{sv}" direction="out"/>
	// 	      </method>
	// 	      <method name="GetObjectDescription">
	// 	         <arg name="objectDescription" type="a(sas)" direction="out"/>
	// 	      </method>
	// 	      <signal name="Announce">
	// 	         <arg name="version" type="q"/>
	// 	         <arg name="port" type="q"/>
	// 	         <arg name="objectDescription" type="a(sas)"/>
	// 	         <arg name="metaData" type="a{sv}"/>
	// 	      </signal>
	// 	   </interface>
	// 	</node>
	// 	`

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

func TestIntrospect(*testing.T) {
	allJoynBridge := NewAllJoynBridge(nil, DummyIntrospectProvider)
	allJoynBridge.AddService("/com/devicehive/alljoyn/test", "com.devicehive.alljoyn.test", "/com/devicehive/alljoyn/test", "com.devicehive.alljoyn.test")
}

func TestGetAllJoynObjects(t *testing.T) {
	node, err := DummyIntrospectProvider("", "")
	if err != nil {
		t.Fail()
	}

	obj := GetAllJoynObjects([]*introspect.Node{node})
	PrintObjects(obj)
}
