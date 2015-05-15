package main

import (
	"log"

	"github.com/devicehive/IoT-framework/godbus-helpers/cloud"
)

func main() {
	c, err := cloud.NewDbusForComDevicehiveCloud()
	if err != nil {
		log.Fatalf("Creation Dbus wrapper with error: %s", err.Error())
	}

	c.SendNotification("[H I G H  N O T I F I C A T I O N]", map[string]interface{}{}, 1000)
}
