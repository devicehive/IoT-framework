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
