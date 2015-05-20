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

	// Optional
	DeviceNotifcationsReceive    string `yaml:"DeviceNotifcationsReceive,omitempty"`
	SendNotificatonQueueCapacity uint64 `yaml:"SendNotificatonQueueCapacity,omitempty"`
}

func (c *Conf) fix() {
	if len(c.DeviceNotifcationsReceive) == 0 {
		c.DeviceNotifcationsReceive = DeviceNotificationReceiveByWS
	}

	if c.SendNotificatonQueueCapacity == 0 {
		c.SendNotificatonQueueCapacity = 2048
	}
}

func FromArgs() (filepath string, c Conf, err error) {
	parseArgs()
	filepath = confArgValue
	if len(filepath) == 0 {
		c = TestConf()
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

	if err == nil {
		(&c).fix()
	}

	return
}

func TestConf() Conf {
	c := Conf{}

	c.URL = "http://52.6.240.235:8080/dh/rest"
	c.AccessKey = "1jwKgLYi/CdfBTI9KByfYxwyQ6HUIEfnGSgakdpFjgk="
	c.DeviceID = "0B24431A-EC99-4887-8B4F-38C3CEAF1D03"
	c.DeviceName = "snappy-go-gateway"

	//c.SendNotificatonQueueCapacity = 23
	// c.DeviceNotifcationsReceive = DeviceNotificationReceiveByREST

	(&c).fix()
	return c
}
