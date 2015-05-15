package ble

import "github.com/devicehive/IoT-framework/godbus-helpers/dbushelper"

type Dbus struct{ *dbushelper.Dbus }

func NewDbus(iface, path string) (*Dbus, error) {
	base, err := dbushelper.NewDbus(iface, path)
	return &Dbus{base}, err
}

func NewDbusForComDevicehiveCloud() (*Dbus, error) {
	return NewDbus(PathComDevicehiveBluetooth, IfaceComDevicehiveBluetooth)
}

const (
	PathComDevicehiveBluetooth  = "/com/devicehive/bluetooth"
	IfaceComDevicehiveBluetooth = "com.devicehive.bluetooth"
)
