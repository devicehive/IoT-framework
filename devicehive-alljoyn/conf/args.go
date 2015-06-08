package conf

import "flag"

const (
	confArgKey          = "conf"
	confArgDefaultValue = ""
)

var (
	confArgValue = ""
)

func init() {
	flag.StringVar(&confArgValue, confArgKey, confArgDefaultValue, "file with DeviceHive configuration in Yaml")
}

func parseArgs() {
	if !flag.Parsed() {
		flag.Parse()
	}
}
