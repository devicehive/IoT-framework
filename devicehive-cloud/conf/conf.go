package conf

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	DeviceNotificationReceiveByWS   = "WS"
	DeviceNotificationReceiveByREST = "REST"
)

type Conf struct {
	URL       string `yaml:"URL,omitempty"`
	AccessKey string `yaml:"AccessKey,omitempty"`

	DeviceID   string `yaml:"DeviceID,omitempty"`
	DeviceName string `yaml:"DeviceName,omitempty"`

	DeviceNotifcationsReceive string `yaml:"DeviceNotifcationsReceive,omitempty"`
}

func defaultConf() Conf {
	return Conf{DeviceNotifcationsReceive: DeviceNotificationReceiveByWS}
}

func FromArgs() (filepath string, c Conf, err error) {
	parseArgs()
	if len(confArgValue) == 0 {
		c = testConf()
		return
	}
	c, err = readConf(confArgValue)
	return
}

func readConf(filepath string) (c Conf, err error) {
	yamlFile, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(yamlFile, &c)
	return
}

func testConf() Conf {
	c := defaultConf()

	c.URL = "http://dh-just-in-case.cloudapp.net:8080/dh/rest"
	c.AccessKey = "1jwKgLYi/CdfBTI9KByfYxwyQ6HUIEfnGSgakdpFjgk="
	c.DeviceID = "0B24431A-EC99-4887-8B4F-38C3CEAF1D03"
	c.DeviceName = "snappy-go-gateway"

	//c.DeviceNotifcationsReceive = DeviceNotificationReceiveByREST

	return c
}
