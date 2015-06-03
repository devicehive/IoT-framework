package main

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

func main() {
	var id uint32 = 782346
	status := "UpdateCommandTestStatus"
	result := "UpdateCommandTestResult"

	f, c, err := conf.FromArgs()
	if err != nil {
		say.Infof("Load conf err: %s", err.Error())
		return
	}

	say.Infof("Conf(%s): %+v", f, c)

	err = rest.DeviceCmdUpdate(c.URL, c.DeviceID, c.AccessKey, id, status, result)

	if err != nil {
		say.Infof("Error: %s", err.Error())
	} else {
		say.Infof("ok")
	}
}
