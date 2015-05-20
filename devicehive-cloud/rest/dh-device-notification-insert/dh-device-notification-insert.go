package main

import (
	"log"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
)

func main() {
	name := "TestRestNotification"
	parameters := map[string]interface{}{"key1": "value1"}

	c := conf.TestConf()

	dnir, err := rest.DeviceNotificationInsert(c.URL, c.DeviceID, c.AccessKey, name, parameters)

	if err != nil {
		log.Printf("Error: %s", err.Error())
	} else {
		log.Printf("Ok: %+v", dnir)
	}
}
