package main

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/godbus/dbus"
)

func restImplementation(bus *dbus.Conn, config conf.Conf) {
	listener := rest.NewDeviceNotificationListener(config.URL, config.DeviceID, func(n rest.DeviceNotification) {
		//TODO: process parameters & send to bus
	})

	listener.Run()
}
