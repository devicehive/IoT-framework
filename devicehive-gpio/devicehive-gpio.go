package main

import (
	"flag"
	"fmt"
	"github.com/godbus/dbus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	BusName = "com.devicehive.gpio"
)

// create new D-Bus connection
func newDBus(name string) (conn *dbus.Conn, err error) {
	// TODO: pass address and select one of system/session/custom

	conn, err = dbus.SystemBus()
	if err != nil {
		return
	}

	reply, err := conn.RequestName(name, dbus.NameFlagDoNotQueue)
	if err != nil {
		return
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		err = fmt.Errorf("D-Bus %q name already taken", name)
	}

	return
}

// create new D-Bus error
func newDBusError(body ...interface{}) *dbus.Error {
	return dbus.NewError(ErrorName, body)
}

// get predefined pin/port dictionary
func loadProfile(profile string) (pins map[string]string, err error) {
	pins = make(map[string]string)

	// check if no profile provided
	if len(profile) == 0 {
		return
	}

	// check profile exists
	info, err := os.Stat(profile)
	if err != nil {
		return pins, fmt.Errorf("failed to access profile %q: %s", profile, err)
	}

	// if directory provided get board model
	if info.IsDir() {
		modelBuf, err := ioutil.ReadFile("/sys/firmware/devicetree/base/model")
		model := strings.TrimSpace(string(modelBuf))
		profile = fmt.Sprintf("%s/%s", profile, model)

		// stat it again
		info, err = os.Stat(profile)
		if err != nil {
			return pins, fmt.Errorf("failed to access profile %q: %s", profile, err)
		}
	}

	// read whole profile
	buf, err := ioutil.ReadFile(profile)
	if err != nil {
		return pins, fmt.Errorf("failed to read profile %q: %s", profile, err)
	}

	// parse YML
	err = yaml.Unmarshal(buf, &pins)
	if err != nil {
		return pins, fmt.Errorf("failed to parse profile %q: %s", profile, err)
	}

	//log.Printf("profile: %+v", pins)
	return
}

// daemon entry point
func main() {
	profile := flag.String("profile", "", "YML file or profile directory")
	flag.Parse()

	// create D-Bus connection
	bus, err := newDBus(BusName)
	if err != nil {
		log.Fatalf("no D-Bus access: %s", err)
	}

	// load profile (might be empty)
	pins, err := loadProfile(*profile)
	if err != nil {
		log.Fatalf("failed to load profile from %q: %s", *profile, err)
	}

	// create service
	s := NewService()
	for pin, port := range pins {
		err = s.createPin(pin, port)
		if err != nil {
			log.Fatalf("failed to create pin:%q, port:%q pair: %s", err)
		}
	}

	// export to D-Bus
	err = s.export(bus)
	if err != nil {
		log.Fatalf("failed to export service: %s", err)
	}

	select {}
}
