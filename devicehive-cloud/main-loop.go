package main

import (
	"encoding/json"

	"github.com/devicehive/IoT-framework/devicehive-cloud/conf"
	"github.com/devicehive/devicehive-go/devicehive"
	"github.com/devicehive/devicehive-go/devicehive/core"
	"github.com/devicehive/devicehive-go/devicehive/log"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"

	"time"
)

const (
	ComDevicehiveCloudPath  = "/com/devicehive/cloud"
	ComDevicehiveCloudIface = "com.devicehive.cloud"

	waitTimeout = 30 * time.Second
)

// DBus wrapper object
type DBusWrapper struct {
	service devicehive.Service
	device  *core.Device
}

// send notification
// priority is ignored
func (w *DBusWrapper) SendNotification(name, parameters string, priority uint64) *dbus.Error {
	log.Infof("sending notification(name=%q, params=%q, priority=%d)", name, parameters, priority)
	dat, err := parseJSON(parameters)
	if err != nil {
		log.Warnf("failed to convert notification parameters to JSON (error: %s)", err)
		return newDHError(err.Error())
	}

	notification := devicehive.NewNotification(name, dat)
	err = w.service.InsertNotification(w.device, notification, waitTimeout)
	if err != nil {
		log.Warnf("failed to send notification (error: %s)", err)
		return newDHError(err.Error())
	}

	return nil // OK
}

// update command result
func (w *DBusWrapper) UpdateCommand(id uint64, status, result string) *dbus.Error {
	log.Infof("updating command(id:%d, status=%q, result:%q", id, status, result)
	dat, err := parseJSON(result)
	if err != nil {
		log.Warnf("failed to convert command result to JSON (error: %s)", err)
		return newDHError(err.Error())
	}

	command := devicehive.NewCommandResult(id, status, dat)
	err = w.service.UpdateCommand(w.device, command, waitTimeout)
	if err != nil {
		log.Warnf("failed to update command (error: %s)", err)
		return newDHError(err.Error())
	}

	return nil // OK
}

// export main + introspectable DBus objects
func exportDBusObject(bus *dbus.Conn, w *DBusWrapper) {
	bus.Export(w, ComDevicehiveCloudPath, ComDevicehiveCloudIface)

	// main service interface
	serviceInterface := introspect.Interface{
		Name:    ComDevicehiveCloudIface,
		Methods: introspect.Methods(w),
		Signals: []introspect.Signal{
			{
				Name: "CommandReceived",
				Args: []introspect.Arg{
					{"id", "t", "in"},
					{"name", "s", "in"},
					{"parameters", "s", "in"}, // JSON string
				},
			},
		},
	}

	// main service node
	n := &introspect.Node{
		Name: ComDevicehiveCloudPath,
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			serviceInterface},
	}
	n_obj := introspect.NewIntrospectable(n)
	log.Tracef("%q introspectable: %s", ComDevicehiveCloudPath, n_obj)
	bus.Export(n_obj, ComDevicehiveCloudPath, "org.freedesktop.DBus.Introspectable")

	// root node
	root := &introspect.Node{
		Children: []introspect.Node{
			{Name: ComDevicehiveCloudPath},
		},
	}
	root_obj := introspect.NewIntrospectable(root)
	log.Tracef("%q introspectable: %s", "/", root_obj)
	bus.Export(root_obj, "/", "org.freedesktop.DBus.Introspectable")
}

// main loop
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

	wrapper := DBusWrapper{service: service, device: device}
	exportDBusObject(bus, &wrapper)

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
			bus.Emit(ComDevicehiveCloudPath, ComDevicehiveCloudIface+".CommandReceived", cmd.Id, cmd.Name, params)
		}

		//time.Sleep(5 * time.Second)
	}
}
