package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"
	dh "github.com/devicehive/devicehive-go"
)

// DBusService is an D-Bus service wrapper.
type DBusService struct {
	service dh.DeviceService // DeviceHive service
	device  *dh.Device       // DeviceHive device
}

// SendNotification sends notification to the DeviceHive device.
// priority is ignored for now.
func (w *DBusService) SendNotification(name, parameters string, priority uint64) *dbus.Error {
	log.WithField("name", name).WithField("params", parameters).
		Debugf("[%s]: sending notification...", TAG)

	// parse JSON parameters
	params, err := parseJSON(parameters)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to convert parameters to JSON", TAG)
		return newDHError(fmt.Sprintf("failed to convert parameters to JSON: %s", err))
	}

	// ensure service is available
	if w.service == nil || w.device == nil {
		// TODO: put to pending queue?
		log.WithError(err).Warnf("[%s]: no DeviceHive service available, ignored", TAG)
		return newDHError("no DeviceHive service available")
	}

	// send notification
	notification := dh.NewNotification(name, params)
	err = w.service.InsertNotification(w.device, notification)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to send notification", TAG)
		return newDHError(fmt.Sprintf("failed to send notification: %s", err))
	}

	log.WithField("name", name).WithField("params", parameters).
		Infof("[%s] notification sent", TAG)

	return nil // OK
}

// UpdateCommand updates the command result.
func (w *DBusService) UpdateCommand(ID uint64, status, result string) *dbus.Error {
	log.WithField("ID", ID).WithField("status", status).WithField("result", result).
		Debugf("[%s]: updating command...", TAG)

	// parse JSON result
	res, err := parseJSON(result)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to convert result to JSON", TAG)
		return newDHError(fmt.Sprintf("failed to convert result to JSON: %s", err))
	}

	// ensure service is available
	if w.service == nil || w.device == nil {
		// TODO: put to pending queue?
		log.WithError(err).Warnf("[%s]: no DeviceHive service available, ignored", TAG)
		return newDHError("no DeviceHive service available")
	}

	// update command
	command := dh.NewCommandResult(ID, status, res)
	err = w.service.UpdateCommand(w.device, command)
	if err != nil {
		log.WithError(err).Warnf("[%s]: failed to update command", TAG)
		return newDHError(fmt.Sprintf("failed to update command: %s", err))
	}

	log.WithField("ID", ID).WithField("status", status).WithField("result", result).
		Infof("[%s]: command updated", TAG)

	return nil // OK
}

// Export exports main D-Bus service and introspectable.
func (w *DBusService) Export(bus *dbus.Conn) error {
	err := bus.Export(w, ComDevicehiveCloudPath, ComDevicehiveCloudIface)
	if err != nil {
		return err
	}

	// main service interface
	serviceIface := introspect.Interface{
		Name:    ComDevicehiveCloudIface,
		Methods: introspect.Methods(w),
		Signals: []introspect.Signal{
			{
				Name: "CommandReceived",
				Args: []introspect.Arg{
					{"id", "t", "out"},
					{"name", "s", "out"},
					{"parameters", "s", "out"}, // JSON string
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
			serviceIface,
		},
	}
	nodeObj := introspect.NewIntrospectable(n)
	log.WithField("node", nodeObj).Debugf("[%s]: %q introspectable", TAG, ComDevicehiveCloudPath)
	err = bus.Export(nodeObj, ComDevicehiveCloudPath, introspect.IntrospectData.Name)
	if err != nil {
		return err
	}

	// root node
	root := &introspect.Node{
		Children: []introspect.Node{
			{Name: strings.TrimPrefix(ComDevicehiveCloudPath, "/")},
		},
	}
	rootObj := introspect.NewIntrospectable(root)
	log.WithField("root", rootObj).Debugf("[%s]: %q introspectable", TAG, "/")
	err = bus.Export(rootObj, "/", introspect.IntrospectData.Name)
	if err != nil {
		return err
	}

	return nil // OK
}

// create new DBus error
func newDHError(message string) *dbus.Error {
	return dbus.NewError("com.devicehive.Error",
		[]interface{}{message})
}

// parse a string to custom JSON
func parseJSON(s string) (dat interface{}, err error) {
	b := bytes.Trim([]byte(s), "\x00")
	err = json.Unmarshal(b, &dat)
	return
}

// format custom JSON to string
func formatJSON(dat interface{}) (s string, err error) {
	if dat == nil {
		return "", nil
	}

	buf, err := json.Marshal(dat)
	if err != nil {
		return "", err
	}

	return string(buf), nil // OK
}
