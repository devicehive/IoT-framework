package main

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"

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
	bus, err := dbus.SystemBus()
	if err != nil {
		log.Printf("Cannot get system bus with error: %s\n", err.Error())
		log.Printf("Trying to use session bus for testing purposes...\n")
		if bus, err = dbus.SessionBus(); err != nil {
			log.Fatalf("Cannot get session bus with error: %s\n", err.Error())
			return
		}
	}

	reply, err := bus.RequestName(DBusConnName, dbus.NameFlagDoNotQueue)
	switch {
	case err != nil:
		log.Fatalf("Cannot request name '%s' with error: %s\n", DBusConnName, err.Error())
	case reply != dbus.RequestNameReplyPrimaryOwner:
		log.Fatalf("The name '%s' already taken.", DBusConnName)
	}

	configFile, config, err := conf.FromArgs()
	println(configFile)
	switch {
	case err != nil:
		log.Fatalf("Cannot read configuration in `%s` with error: %s", configFile, err.Error())
	case configFile == "":
		log.Printf("You should specify configuration file.\n Starting with test configuration: %+v", config)
	default:
		log.Printf("Starting DeviceHive gateway with configuration in '%s': %+v", configFile, config)
	}

	if config.DeviceNotifcationsReceive == conf.DeviceNotificationReceiveByWS {
		wsImplementation(bus, config)
		return
	}

	if config.DeviceNotifcationsReceive == conf.DeviceNotificationReceiveByREST {
		restImplementation(bus, config)
		return
	}

}
