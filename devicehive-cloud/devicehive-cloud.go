package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/godbus/dbus"
	dh "gopkg.in/devicehive/devicehive-go.v2"
)

const (
	ComDevicehiveCloudPath  = "/com/devicehive/cloud"
	ComDevicehiveCloudIface = "com.devicehive.cloud"
	DBusConnName            = "com.devicehive.cloud"

	DefaultWaitTimeout = 45 * time.Second
	RetryTimeout       = 5 * time.Second

	DeviceClassName    = "go-gateway"
	DeviceClassVersion = "0.2"
)

var (
	// package logger instance
	log = logrus.New()

	// TAG is a log prefix
	TAG = "DH"
)

var (
	lastTimestamp = ""

	// service configuration
	config Config
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("[%s]: %v", TAG, r)
		}
	}()
	parseArgs()

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
	// getting DeviceHive service
	log.Debugf("[%s]: creating DeviceHive service...", TAG)
	service, err := newDeviceService(config.URL, config.AccessKey, config.DeviceHiveLogLevel)
	if err != nil {
		return fmt.Errorf("failed to create DeviceHive service: %s", err)
	}
	service.SetTimeout(DefaultWaitTimeout)
	log.WithField("service", service).Infof("[%s]: DeviceHive service used", TAG)
	defer service.Stop()

	// getting server info
	if len(lastTimestamp) == 0 {
		log.Debugf("[%s]: getting DeviceHive service info...", TAG)
		info, err := service.GetServerInfo()
		if err != nil {
			return fmt.Errorf("failed to get service info: %s", err)
		}
		log.WithField("info", info).Infof("[%s]: DeviceHive service info", TAG)
		lastTimestamp = info.Timestamp
	}

	// registering device
	log.Debugf("[%s]: registering gateway device...", TAG)
	device := dh.NewDevice(config.DeviceID, config.DeviceName,
		dh.NewDeviceClass(DeviceClassName, DeviceClassVersion))
	device.Key = config.DeviceKey
	if len(config.NetworkName) != 0 || len(config.NetworkKey) != 0 {
		device.Network = dh.NewNetwork(config.NetworkName, config.NetworkKey)
		device.Network.Description = config.NetworkDesc
	}
	err = service.RegisterDevice(device)
	if err != nil {
		return fmt.Errorf("failed to register device: %s", err)
	}
	log.WithField("device", device).Infof("[%s]: gateway device registered", TAG)

	// start polling commands
	log.Debugf("[%s]: subscribing gateway device commands...", TAG)
	listener, err := service.SubscribeCommands(device, lastTimestamp)
	if err != nil {
		return fmt.Errorf("failed to subscribe commands: %s", err)
	}

	// D-Bus service wrapper
	log.Debugf("[%s]: exporting D-Bus service...", TAG)
	wrapper := newDBusService(bus)
	wrapper.service = service
	wrapper.device = device
	err = wrapper.export()
	if err != nil {
		return fmt.Errorf("failed to export D-Bus service: %s", err)
	}

	// main loop
	log.Debugf("[%s]: running...", TAG)
	for {
		select {
		case command := <-listener.C:
			if command != nil {
				log.WithField("command", command).Infof("[%s]: new command received", TAG)
				params, err := formatJSON(command.Parameters)
				if err != nil {
					log.WithError(err).Warnf("[%s]: failed to generate JSON from parameters, ignored", TAG)
					continue
				}

				if err = wrapper.emitCommandReceived(command.ID, command.Name, params); err != nil {
					log.WithError(err).Warnf("[%s]: failed to emit D-Bus signal, ignored", TAG)
					continue
				}
			}

		case <-stopCh:
			return nil // stop!
		}
	}
}

// newDeviceService creates a new device service (either REST or Websocket).
// If protocol is "ws://" or "wss://" Websocket service will be created,
// otherwise REST service will be used as a fallback.
// Access token is important!
func newDeviceService(baseURL, accessToken, logLevel string) (dh.DeviceService, error) {
	url := strings.ToLower(baseURL)
	if strings.HasPrefix(url, `ws://`) || strings.HasPrefix(url, `wss://`) {
		if len(logLevel) != 0 {
			_ = dh.SetLogLevel(logLevel)
		}
		return dh.NewWsDeviceService(baseURL, accessToken)
	}

	// use REST service as a fallback
	if len(logLevel) != 0 {
		_ = dh.SetLogLevel(logLevel)
	}
	return dh.NewRestService(baseURL, accessToken)
}

// parse command line arguments
func parseArgs() {
	if !flag.Parsed() {
		flag.Parse()
	}

	if len(configFileName) == 0 {
		// no file provided
		flag.Usage()
		os.Exit(1)
	}

	if err := config.FromFile(configFileName); err != nil {
		panic(err) // failed to parse configuration
	}

	log.Level, _ = logrus.ParseLevel(config.LoggingLevel)
}
