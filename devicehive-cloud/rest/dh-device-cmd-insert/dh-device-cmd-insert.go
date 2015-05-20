package main

import (
	"log"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
)

func main() {
	command := "TestRestCommand"
	parameters := map[string]interface{}{"key1": "value1"}

	c := conf.TestConf()

	dcir, err := rest.DeviceCmdInsert(c.URL, c.DeviceID, c.AccessKey, command, parameters)

	if err != nil {
		log.Printf("Error: %s", err.Error())
	} else {
		log.Printf("Ok: %+v", dcir)
	}
}
