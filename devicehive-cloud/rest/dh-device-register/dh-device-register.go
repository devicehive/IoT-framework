package main

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

func main() {
	say.Level = say.DEBUG

	f, c, err := conf.FromArgs()
	if err != nil {
		say.Infof("Load conf err: %s", err.Error())
		return
	}

	say.Infof("Conf(%s): %+v", f, c)

	err = rest.DeviceRegisterEasy(c.URL, c.DeviceID, c.AccessKey, c.DeviceName)

	if err != nil {
		say.Infof("Error: %s", err.Error())
	} else {
		say.Infof("Ok")
	}
}
