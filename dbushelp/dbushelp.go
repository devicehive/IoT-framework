package dbushelp

import (
	"encoding/json"

	"github.com/godbus/dbus"
)

// Helpers for github.com/godbus/dbus

// Dbus is a *dbus.Conn wrapper
type Dbus struct {
	conn *dbus.Conn

	// See example constants
	Path, Iface string
}

func NewDbus() (*Dbus, error) {
	w := Dbus{}

	conn, err := dbus.SystemBus()
	if err != nil {
		conn, err = dbus.SessionBus()
		if err != nil {
			return nil, err
		}
	}

	w.conn = conn
	return &w, nil
}

func (w *Dbus) Conn() *dbus.Conn {
	return w.conn
}

func (w *Dbus) Call(name string, args ...interface{}) *dbus.Call {
	c := w.conn.Object(w.Iface, dbus.ObjectPath(w.Path)).Call(w.Iface+"."+name, 0, args...)
	return c
}

func (w *Dbus) SendNotification(name, parameters interface{}, priority uint64) (*dbus.Call, error) {
	b, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	c := w.Call("SendNotification", name, string(b), priority)
	return c, c.Err
}

//TODO: Adding new methods if necessary
