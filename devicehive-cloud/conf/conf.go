package conf

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	URL       string `yaml:"URL,omitempty"`
	AccessKey string `yaml:"AccessKey,omitempty"`

	DeviceID   string `yaml:"DeviceID,omitempty"`
	DeviceName string `yaml:"DeviceName,omitempty"`
	DeviceKey  string `yaml:"DeviceKey,omitempty"`

	NetworkName string `yaml:"NetworkName,omitempty"`
	NetworkKey  string `yaml:"NetworkKey,omitempty"`
	NetworkDesc string `yaml:"NetworkDescription,omitempty"`

	// Optional
	SendNotificatonQueueCapacity uint64 `yaml:"SendNotificatonQueueCapacity,omitempty"`
	LoggingLevel                 string `yaml:"LoggingLevel,omitempty"`
}

func (c *Conf) fix() {
	if c.SendNotificatonQueueCapacity == 0 {
		c.SendNotificatonQueueCapacity = 2048
	}

	if len(c.LoggingLevel) == 0 {
		c.LoggingLevel = "info"
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
	c.DeviceKey = "snappy-go-secret-key"

	// c.LoggingLevel = "info"
	// c.LoggingLevel = "debug"
	c.LoggingLevel = "trace"

	(&c).fix()
	return c
}
