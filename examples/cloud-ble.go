package main

import (
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus"
	"log"
	"strings"
	"sync"
)

type dbusWrapper struct {
	conn         *dbus.Conn
	path, iface  string
	handlers     map[string]signalHandler
	handlersSync sync.Mutex
}

type signalHandler func(args ...interface{})
type cloudCommandHandler func(map[string]interface{}) (map[string]interface{}, error)

func NewdbusWrapper(path string, iface string) (*dbusWrapper, error) {
	d := new(dbusWrapper)

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Panic(err)
	}

	d.handlers = make(map[string]signalHandler)

	d.conn = conn
	d.path = path
	d.iface = iface

	filter := fmt.Sprintf("type='signal',path='%[1]s',interface='%[2]s',sender='%[2]s'", path, iface)
	log.Printf("Filter: %s", filter)

	conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").Call("org.freedesktop.DBus.AddMatch", 0, filter)

	go func() {
		ch := make(chan *dbus.Signal, 100)
		conn.Signal(ch)
		for signal := range ch {
			if !((strings.Index(signal.Name, iface) == 0) && (string(signal.Path) == path)) {
				continue
			}

			log.Printf("Received: %s", signal)
			if val, ok := d.handlers[signal.Name]; ok {
				val(signal.Body...)
			}
		}
	}()

	return d, nil
}

func (d *dbusWrapper) call(name string, args ...interface{}) *dbus.Call {
	c := d.conn.Object(d.iface, dbus.ObjectPath(d.path)).Call(d.iface+"."+name, 0, args...)

	if c.Err != nil {
		log.Printf("Error calling %s: %s", name, c.Err)
	}

	return c
}

func (d *dbusWrapper) SendNotification(name string, parameters interface{}) {
	b, _ := json.Marshal(parameters)
	d.call("SendNotification", name, string(b))
}

func (d *dbusWrapper) RegisterHandler(signal string, h signalHandler) {
	d.handlersSync.Lock()
	d.handlers[d.iface+"."+signal] = h
	d.handlersSync.Unlock()
}

func (d *dbusWrapper) BleScanStart() error {
	c := d.call("ScanStart")
	return c.Err
}

func (d *dbusWrapper) BleConnect(mac string) error {
	c := d.call("Connect", mac)
	return c.Err
}

func (d *dbusWrapper) BleScanStop() error {
	c := d.call("ScanStop")
	return c.Err
}

func (d *dbusWrapper) BleGattRead(mac, uuid string) (map[string]interface{}, error) {
	s := ""
	err := d.call("GattRead", mac, uuid).Store(&s)

	res := map[string]interface{}{
		"value": s,
	}

	return res, err
}

func (d *dbusWrapper) BleGattWrite(mac, uuid, message string) (map[string]interface{}, error) {
	c := d.call("GattWrite", mac, uuid, message)
	return nil, c.Err
}

func (d *dbusWrapper) BleGattNotifications(mac, uuid string, enable bool) (map[string]interface{}, error) {
	c := d.call("GattNotifications", mac, uuid, enable)
	return nil, c.Err
}

func (d *dbusWrapper) BleGattIndications(mac, uuid string, enable bool) (map[string]interface{}, error) {
	c := d.call("GattIndications", mac, uuid, enable)
	return nil, c.Err
}

func (d *dbusWrapper) CloudUpdateCommand(id uint32, status string, result map[string]interface{}) {
	b, _ := json.Marshal(result)
	d.call("UpdateCommand", id, status, string(b))
}

func main() {
	cloud, err := NewdbusWrapper("/com/devicehive/cloud", "com.devicehive.cloud")
	if err != nil {
		log.Panic(err)
	}

	ble, err := NewdbusWrapper("/com/devicehive/bluetooth", "com.devicehive.bluetooth")
	if err != nil {
		log.Panic(err)
	}

	ble.RegisterHandler("DeviceDiscovered", func(args ...interface{}) {
		cloud.SendNotification("PeripheralDiscovered", map[string]interface{}{
			"mac":  args[0].(string),
			"name": args[1].(string),
			"rssi": args[2].(int16),
		})
	})

	ble.RegisterHandler("DeviceConnected", func(args ...interface{}) {
		cloud.SendNotification("DeviceConnected", map[string]interface{}{
			"mac": args[0].(string),
		})
	})

	ble.RegisterHandler("PeripheralConnected", func(args ...interface{}) {
		cloud.SendNotification("PeripheralConnected", map[string]interface{}{
			"mac": args[0].(string),
		})
	})

	ble.RegisterHandler("NotificationReceived", func(args ...interface{}) {
		cloud.SendNotification("NotificationReceived", map[string]interface{}{
			"mac":            args[0].(string),
			"characteristic": args[1].(string),
			"value":          args[2].(string),
		})
	})

	ble.RegisterHandler("IndicationReceived", func(args ...interface{}) {
		cloud.SendNotification("IndicationReceived", map[string]interface{}{
			"mac":            args[0].(string),
			"characteristic": args[1].(string),
			"value":          args[2].(string),
		})
	})

	cloudHandlers := make(map[string]cloudCommandHandler)

	cloudHandlers["connect"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		return nil, ble.BleConnect(p["mac"].(string))
	}

	cloudHandlers["scan/start"] = func(map[string]interface{}) (map[string]interface{}, error) {
		return nil, ble.BleScanStart()
	}

	cloudHandlers["scan/stop"] = func(map[string]interface{}) (map[string]interface{}, error) {
		ble.BleScanStop()
		return nil, ble.BleScanStop()
	}

	cloudHandlers["gatt/read"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		return ble.BleGattRead(p["mac"].(string), p["uuid"].(string))
	}

	cloudHandlers["gatt/write"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattWrite(p["mac"].(string), p["uuid"].(string), p["value"].(string))
		return nil, err
	}

	cloudHandlers["gatt/notifications"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattNotifications(p["mac"].(string), p["uuid"].(string), true)
		return nil, err
	}

	cloudHandlers["gatt/notifications/stop"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattNotifications(p["mac"].(string), p["uuid"].(string), false)
		return nil, err
	}

	cloudHandlers["gatt/indications"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattIndications(p["mac"].(string), p["uuid"].(string), true)
		return nil, err
	}

	cloudHandlers["gatt/indications/stop"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		_, err := ble.BleGattIndications(p["mac"].(string), p["uuid"].(string), false)
		return nil, err
	}

	cloud.RegisterHandler("CommandReceived", func(args ...interface{}) {
		id := args[0].(uint32)
		command := args[1].(string)
		params := args[2].(string)

		var dat map[string]interface{}
		b := []byte(params)
		json.Unmarshal(b, &dat)

		if h, ok := cloudHandlers[command]; ok {
			res, err := h(dat)

			if err != nil {
				cloud.CloudUpdateCommand(id, fmt.Sprintf("ERROR: %s", err.Error()), nil)
			} else {
				cloud.CloudUpdateCommand(id, "OK", res)
			}

		} else {
			log.Printf("Unhandled command: %s", command)
		}
	})

	select {}
}
