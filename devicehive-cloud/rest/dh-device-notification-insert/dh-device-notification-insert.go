package main

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

func main() {
	name := "TestRestNotification"
	parameters := map[string]interface{}{"key1": "value1"}

	f, c, err := conf.FromArgs()
	if err != nil {
		say.Infof("Load conf err: %s", err.Error())
		return
	}

	say.Infof("Conf(%s): %+v", f, c)

	dnir, err := rest.DeviceNotificationInsert(c.URL, c.DeviceID, c.AccessKey, name, parameters)

	if err != nil {
		say.Infof("Error: %s", err.Error())
	} else {
		say.Infof("Ok: %+v", dnir)
	}
}
