package conf

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

const (
	argKey     = "conf"
	defaultVal = ""
)

type Conf struct {
	URL       string
	AccessKey string

	DeviceID   string
	DeviceName string
}

func FromArgs() (filepath string, c Conf, err error) {
	pFilePath := flag.String(argKey, defaultVal, "file with DeviceHive configuration in Yaml")
	filepath = *pFilePath
	if len(filepath) == 0 {
		c = testConf()
		return
	}
	c, err = readConf(filepath)
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
