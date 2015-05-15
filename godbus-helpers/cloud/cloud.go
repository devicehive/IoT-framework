package cloud

import (
	"encoding/json"

	"github.com/godbus/dbus"

	"github.com/devicehive/IoT-framework/godbus-helpers/dbushelper"
)

type Dbus struct{ *dbushelper.Dbus }

func NewDbus(iface, path string) (*Dbus, error) {
	base, err := dbushelper.NewDbus(iface, path)
	return &Dbus{base}, err
}

func NewDbusForComDevicehiveCloud() (*Dbus, error) {
	return NewDbus(PathComDevicehiveCloud, IfaceComDevicehiveCloud)
}

func (w *Dbus) SendNotification(name, parameters interface{}, priority uint64) (*dbus.Call, error) {
	b, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	c := w.Call("SendNotification", name, string(b), priority)
	return c, c.Err
}

const (
	PathComDevicehiveCloud  = "/com/devicehive/cloud"
	IfaceComDevicehiveCloud = "com.devicehive.cloud"
)
