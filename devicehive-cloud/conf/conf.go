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
	return Conf{
		URL:       "http://52.1.250.210:8080/dh/rest",
		AccessKey: "1jwKgLYi/CdfBTI9KByfYxwyQ6HUIEfnGSgakdpFjgk=",

		DeviceID:   "0B24431A-EC99-4887-8B4F-38C3CEAF1D03",
		DeviceName: "snappy-go-gateway",
	}
}
