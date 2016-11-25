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

package dbushelper

import "github.com/godbus/dbus"

// Helpers for github.com/godbus/dbus

type Dbus struct {
	conn        *dbus.Conn
	path, iface string
}

func (w *Dbus) Conn() *dbus.Conn { return w.conn }
func (w *Dbus) Path() string     { return w.path }
func (w *Dbus) Iface() string    { return w.iface }

func NewDbus(path, iface string) (*Dbus, error) {
	w := new(Dbus)

	conn, err := dbus.SystemBus()
	if err != nil {
		conn, err = dbus.SessionBus()
		if err != nil {
			return nil, err
		}
	}

	w.path = path
	w.iface = iface
	w.conn = conn
	return w, nil
}

func (w *Dbus) Call(name string, args ...interface{}) *dbus.Call {
	c := w.conn.Object(w.Iface(), dbus.ObjectPath(w.Path())).Call(w.Iface()+"."+name, 0, args...)
	return c
}
