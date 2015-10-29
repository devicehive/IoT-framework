package main

import (
	"encoding/json"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/devicehive-go/devicehive"
	"github.com/devicehive/devicehive-go/devicehive/log"

	"github.com/godbus/dbus"
	//"github.com/godbus/dbus/introspect"
	//"github.com/godbus/dbus/prop"

	"time"
)

const (
	dbusObjectPath  = "/com/devicehive/cloud"
	dbusCommandName = "com.devicehive.cloud.CommandReceived"

	waitTimeout = 30 * time.Second
)

func mainLoop(bus *dbus.Conn, service devicehive.Service, config conf.Conf) {
	// getting server info
	info, err := service.GetServerInfo(waitTimeout)
	if err != nil {
		log.Warnf("Cannot get service info (error: %s)", err)
		return
	}

	// registering device
	device := devicehive.NewDevice(config.DeviceID, config.DeviceName,
		devicehive.NewDeviceClass("go-gateway-class", "0.1"))
	err = service.RegisterDevice(device, waitTimeout)
	if err != nil {
		log.Warnf("Cannot register device (error: %s)", err)
		return
	}

	// start polling commands
	listener, err := service.SubscribeCommands(device, info.Timestamp, waitTimeout)
	if err != nil {
		log.Warnf("Cannot subscribe commands (error: %s)")
		return
	}

	//	go func() {
	//		nControl := rest.NewPollAsync()
	//		cControl := rest.NewPollAsync()
	//		nOut := make(chan rest.DeviceNotificationResource, 16)
	//		cOut := make(chan rest.DeviceCmdResource, 16)

	//		go rest.DeviceNotificationPollAsync(config.URL, config.DeviceID, config.AccessKey, nOut, nControl)
	//		go rest.DeviceCmdPollAsync(config.URL, config.DeviceID, config.AccessKey, cOut, cControl)

	//		for {
	//			select {
	//			case n := <-nOut:
	//				parameters := ""
	//				if n.Parameters != nil {
	//					b, err := json.Marshal(n.Parameters)
	//					if err != nil {
	//						say.Infof("Could not generate JSON from parameters of notification %+v\nWith error %s", n, err.Error())
	//						continue
	//					}

	//					parameters = string(b)
	//				}
	//				say.Debugf("NOTIFICATION %s -> %s(%v)", config.URL, n.Notification, parameters)
	//				bus.Emit(restObjectPath, restCommandName, uint32(n.Id), n.Notification, parameters)
	//			case c := <-cOut:
	//				parameters := ""
	//				if c.Parameters != nil {
	//					b, err := json.Marshal(c.Parameters)
	//					if err != nil {
	//						say.Infof("Could not generate JSON from parameters of command %+v\nWith error %s", c, err.Error())
	//						continue
	//					}

	//					parameters = string(b)

	//				}
	//				say.Debugf("COMMAND %s -> %s(%v)", config.URL, c.Command, parameters)
	//				bus.Emit(restObjectPath, restCommandName, uint32(c.Id), c.Command, parameters)
	//			}
	//		}
	//	}()

	bus.Export(service, "/com/devicehive/cloud", DBusConnName)

	// Introspectable
	//	n := &introspect.Node{
	//		Name: "/com/devicehive/cloud",
	//		Interfaces: []introspect.Interface{
	//			introspect.IntrospectData,
	//			prop.IntrospectData,
	//			{
	//				Name:    "com.devicehive.cloud",
	//				Methods: introspect.Methods(w),
	//			},
	//		},
	//	}

	//	bus.Export(introspect.NewIntrospectable(n), "/com/devicehive/cloud", "org.freedesktop.DBus.Introspectable")

	for {
		select {
		case cmd := <-listener.C:
			params := ""
			if cmd.Parameters != nil {
				buf, err := json.Marshal(cmd.Parameters)
				if err != nil {
					log.Warnf("Cannot generate JSON from parameters of command %+v (error: %s)", cmd, err)
					continue
				}
				params = string(buf)
			}
			log.Infof("COMMAND %s -> %s(%v)", config.URL, cmd.Name, params)
			bus.Emit(dbusObjectPath, dbusCommandName, uint32(cmd.Id), cmd.Name, params)
		}

		time.Sleep(5 * time.Second)
	}
}
