package main

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

func main() {
	say.Level = say.DEBUG

	c := conf.TestConf()
	err := rest.DeviceRegisterEasy(c.URL, c.DeviceID, c.DeviceName, c.AccessKey)

	if err != nil {
		say.Infof("Error: %s", err.Error())
	} else {
		say.Infof("Ok")
	}
}
