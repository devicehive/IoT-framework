package main

import (
	"log"
	"time"

	"github.com/devicehive/IoT-framework/godbus-helpers/cloud"
)

func main() {
	c, err := cloud.NewDbusForComDevicehiveCloud()
	if err != nil {
		log.Fatalf("Creation Dbus wrapper with error: %s", err.Error())
	}

	for {
		c.SendNotification("TestNotification", map[string]interface{}{}, 1)
		time.Sleep(1 * time.Second)
	}
}
