package main

import (
	"log"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/rest"
)

func main1() {
	c := conf.TestConf()

	dcr, err := rest.DeviceCmdPoll(c.URL, c.DeviceID, c.AccessKey, nil, nil)

	if err != nil {
		log.Printf("Error: %s", err.Error())
	} else {
		log.Printf("Ok: %+v", dcr)
	}

}

func main2() {
	c := conf.TestConf()

	control := rest.NewPollAsync()
	out := make(chan rest.DeviceCmdResource, 16)

	go rest.DeviceCmdPollAsync(c.URL, c.DeviceID, c.AccessKey, out, control)

	for {
		select {
		case item := <-out:
			log.Printf("*item: %+v", item)
			// case <-time.After(15 * time.Second):
			// 	log.Printf("*start Stop()")
			// 	control.Stop()
			// 	log.Printf("*finish Stop()")
			// 	return
		}
	}
}

func main() {
	// main1()
	main2()
}
