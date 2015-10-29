package main

import (
	"bytes"
	"encoding/json"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/devicehive-go/devicehive"
	"github.com/devicehive/devicehive-go/devicehive/log"

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
	configFile, config, err := conf.FromArgs()
	switch {
	case err != nil:
		log.Fatalf("Failed to read %q configuration (%s)", configFile, err)
	case configFile == "":
		log.Warnf("No configuration file provided!")
		log.Infof("Test configuration is used: %+v", config)
	default:
		log.Infof("Starting DeviceHive with %q configuration: %+v", configFile, config)
	}

	log.SetLevelByName(config.LoggingLevel)

	bus, err := dbus.SystemBus()
	if err != nil {
		log.Warnf("Cannot get system bus with error: %s", err)
		log.Infof("Trying to use session bus for testing purposes...")
		if bus, err = dbus.SessionBus(); err != nil {
			log.Fatalf("Cannot get session bus with error: %s", err)
			return
		}
	}

	reply, err := bus.RequestName(DBusConnName, dbus.NameFlagDoNotQueue)
	switch {
	case err != nil:
		log.Fatalf("Cannot request name %q with error: %s", DBusConnName, err)
	case reply != dbus.RequestNameReplyPrimaryOwner:
		log.Fatalf("The name %q already taken", DBusConnName)
	}

	s, err := devicehive.NewService(config.URL, config.AccessKey)
	if err != nil {
		log.Fatalf("Failed to create DeviceHive service (error: %s)", err)
	}
	log.Infof("Starting %v", s)
	mainLoop(bus, s, config)
}
