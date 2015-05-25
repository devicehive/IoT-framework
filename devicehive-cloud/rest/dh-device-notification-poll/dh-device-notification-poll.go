package main

import (
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

func mainEasyTest() {
	c := conf.TestConf()

	dcr, err := rest.DeviceNotificationPoll(c.URL, c.DeviceID, c.AccessKey, nil, nil, nil)

	if err != nil {
		say.Infof("Error: %s", err.Error())
	} else {
		say.Infof("Ok: %+v", dcr)
	}
}

func mainInfinityLoop() {
	c := conf.TestConf()

	control := rest.NewPollAsync()
	out := make(chan rest.DeviceNotificationResource, 16)

	go rest.DeviceNotificationPollAsync(c.URL, c.DeviceID, c.AccessKey, "", out, control)

	for {
		select {
		case item := <-out:
			say.Infof("item: %+v", item)
		}
	}
}

func mainInfinityLoopWithInterruption() {
	c := conf.TestConf()

	control := rest.NewPollAsync()
	out := make(chan rest.DeviceNotificationResource, 16)

	go rest.DeviceNotificationPollAsync(c.URL, c.DeviceID, c.AccessKey, "", out, control)

	for {
		select {
		case item := <-out:
			say.Infof("item: %+v", item)
		case <-time.After(15 * time.Second):
			say.Infof("start Stop()")
			control.Stop()
			say.Infof("finish Stop()")
			return
		}
	}
}

func main() {
	say.Level = say.DEBUG
	say.Infof("POLLING NOTIFICATIONS TEST. Send notification from another terminal")

	mainInfinityLoop()
	//Choose another main function
}
