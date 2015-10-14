package main

import (
	"bytes"
	"strings"
	"encoding/json"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"

	"github.com/godbus/dbus"
)

func parseJSON(s string) (map[string]interface{}, error) {
	var dat map[string]interface{}
	b := []byte(s)
	b = bytes.Trim(b, "\x00")
	err := json.Unmarshal(b, &dat)

	return dat, err
}

func newDHError(message string) *dbus.Error {
	return dbus.NewError("com.devicehive.Error",
		[]interface{}{message})
}

const (
	DBusConnName = "com.devicehive.cloud"
)

func main() {
	say.SetLevelWithConfName("info")

	configFile, config, err := conf.FromArgs()
	switch {
	case err != nil:
		say.Fatalf("Cannot read configuration in `%s` with error: %s", configFile, err.Error())
	case configFile == "":
		say.Infof("You should specify configuration file. Starting with test configuration: %+v", config)
	default:
		say.Infof("Starting DeviceHive gateway with configuration in '%s': %+v", configFile, config)
	}

	say.SetLevelWithConfName(config.LoggingLevel)

	bus, err := dbus.SystemBus()
	if err != nil {
		say.Infof("Cannot get system bus with error: %s", err.Error())
		say.Infof("Trying to use session bus for testing purposes...")
		if bus, err = dbus.SessionBus(); err != nil {
			say.Fatalf("Cannot get session bus with error: %s\n", err.Error())
			return
		}
	}

	reply, err := bus.RequestName(DBusConnName, dbus.NameFlagDoNotQueue)
	switch {
	case err != nil:
		say.Fatalf("Cannot request name '%s' with error: %s\n", DBusConnName, err.Error())
	case reply != dbus.RequestNameReplyPrimaryOwner:
		say.Fatalf("The name '%s' already taken.", DBusConnName)
	}

	switch strings.ToUpper(config.DeviceNotifcationsReceive) {
		case conf.DeviceNotificationReceiveByWS: {
			say.Infof("Starting as websocket...")
			wsImplementation(bus, config)
		}

		case conf.DeviceNotificationReceiveByREST: {
			say.Infof("Starting as rest...")
			restImplementation(bus, config)
		}

		default: {
			say.Fatalf("unknown implementation %q", config.DeviceNotifcationsReceive)
		}
	}
}
