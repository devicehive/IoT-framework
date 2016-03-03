package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/godbus/dbus"
)

const (
	ComDevicehiveBluetoothPath  = "/com/devicehive/bluetooth"
	ComDevicehiveBluetoothIface = "com.devicehive.bluetooth"
	DBusConnName                = "com.devicehive.bluetooth"

	RetryTimeout   = 5 * time.Second
	ConnectTimeout = 3 * time.Second // Peripheral connect timeout
	ExploreTimeout = 5 * time.Second // Timeout to explore peripherals for a newly found device
)

var (
	// package logger instance
	log = logrus.New()

	// TAG is a log prefix
	TAG = "BLE"
)

var (
	deviceID int = 0
)

// populate flags
func init() {
	flag.IntVar(&deviceID, "device-id", 0, "HCI device index")
}

// entry point
func main() {
	// recover from panic
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("[%s]: %v", TAG, r)
		}
	}()

	// parse command line
	if !flag.Parsed() {
		flag.Parse()
	}

	// getting D-Bus bus
	bus, err := dbus.SystemBus()
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to get system bus", TAG)
		log.Infof("trying to use session bus...")
		if bus, err = dbus.SessionBus(); err != nil {
			log.WithError(err).Fatalf("[%s]: failed to get session bus", TAG)
		}
	}

	// request name
	reply, err := bus.RequestName(DBusConnName, dbus.NameFlagDoNotQueue)
	switch {
	case err != nil:
		log.WithError(err).Fatalf("[%s]: failed to request name %q", TAG, DBusConnName)
	case reply != dbus.RequestNameReplyPrimaryOwner:
		log.Fatalf("[%s] the %q name already taken", TAG, DBusConnName)
	}

	log.Infof("[%s] starting...", TAG)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, os.Kill)
	for {
		err := mainLoop(bus, sigCh)
		if err != nil {
			log.WithError(err).Warnf("[%s]: failed", TAG)
			select {
			case <-time.After(RetryTimeout):
				log.Infof("[%s]: try again...", TAG)
				continue

			case <-sigCh:
				// break
			}
		}

		log.Infof("[%s]: stopped", TAG)
		return // full stop
	}
}

// main loop
func mainLoop(bus *dbus.Conn, stopCh chan os.Signal) error {
	// D-Bus service wrapper
	wrapper, err := newDBusService(bus, deviceID)
	if err != nil {
		return fmt.Errorf("failed to create D-Bus service: %s", err)
	}

	log.Debugf("[%s]: exporting D-Bus service...", TAG)
	if err := wrapper.export(); err != nil {
		return fmt.Errorf("failed to export D-Bus service: %s", err)
	}

	// main loop
	log.Debugf("[%s]: running...", TAG)
	for {
		select {
		case <-stopCh:
			return nil // stop!
		}
	}
}
