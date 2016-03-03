package main

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/devicehive/gatt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

// DBusService is a D-Bus service wrapper.
type DBusService struct {
	bus *dbus.Conn

	device    gatt.Device
	connected bool

	devicesDiscovered map[string]*DeviceInfo
	devicesConnected  map[string]*DeviceInfo
	devicesLock       sync.Mutex

	connChan   chan bool
	connecting bool
}

// DeviceInfo is a discovered or/and connected device info.
type DeviceInfo struct {
	name string
	rssi int

	peripheral      gatt.Peripheral
	characteristics map[string]*gatt.Characteristic

	ready         bool
	connectedOnce bool

	lock sync.Mutex
}

// create new D-Bus service object.
func newDBusService(bus *dbus.Conn, deviceID int) (*DBusService, error) {
	d, err := gatt.NewDevice(gatt.LnxDeviceID(deviceID, false))
	if err != nil {
		return nil, fmt.Errorf("Failed to open device %d: %s", deviceID, err)
	}

	w := new(DBusService)
	w.bus = bus
	w.devicesDiscovered = make(map[string]*DeviceInfo)
	w.devicesConnected = make(map[string]*DeviceInfo)
	w.connChan = make(chan bool, 1)

	d.Handle(gatt.PeripheralDiscovered(w.onPeripheralDiscovered))
	d.Handle(gatt.PeripheralConnected(w.onPeripheralConnected))
	d.Handle(gatt.PeripheralDisconnected(w.onPeripheralDisconnected))
	d.Init(w.onInit)
	w.device = d

	return w, nil // OK
}

// export exports main D-Bus service and introspectable.
func (w *DBusService) export() error {
	err := w.bus.Export(w, ComDevicehiveBluetoothPath, ComDevicehiveBluetoothIface)
	if err != nil {
		return err
	}

	// main service node
	n := &introspect.Node{
		Interfaces: []introspect.Interface{
			{
				Name:    ComDevicehiveBluetoothIface,
				Methods: introspect.Methods(w),
				Signals: []introspect.Signal{
					introspect.Signal{
						Name: "PeripheralDiscovered",
						Args: []introspect.Arg{
							{"id", "s", "out"},
							{"name", "s", "out"},
							{"rssi", "i", "out"},
						},
					},
					introspect.Signal{
						Name: "PeripheralConnected",
						Args: []introspect.Arg{
							{"id", "s", "out"},
						},
					},
					introspect.Signal{
						Name: "PeripheralDisconnected",
						Args: []introspect.Arg{
							{"id", "s", "out"},
						},
					},
					introspect.Signal{
						Name: "NotificationReceived",
						Args: []introspect.Arg{
							{"mac", "s", "out"},
							{"uuid", "s", "out"},
							{"value", "s", "out"},
						},
					},
					introspect.Signal{
						Name: "IndicationReceived",
						Args: []introspect.Arg{
							{"mac", "s", "out"},
							{"uuid", "s", "out"},
							{"value", "s", "out"},
						},
					},
				},
			},
		},
	}

	nodeObj := introspect.NewIntrospectable(n)
	log.WithField("node", nodeObj).Debugf("[%s]: %q introspectable", TAG, ComDevicehiveBluetoothPath)
	err = w.bus.Export(nodeObj, ComDevicehiveBluetoothPath, introspect.IntrospectData.Name)
	if err != nil {
		return err
	}

	// root node
	root := &introspect.Node{
		Children: []introspect.Node{
			{Name: strings.TrimPrefix(ComDevicehiveBluetoothPath, "/")},
		},
	}
	rootObj := introspect.NewIntrospectable(root)
	log.WithField("root", rootObj).Debugf("[%s]: %q introspectable", TAG, "/")
	err = w.bus.Export(rootObj, "/", introspect.IntrospectData.Name)
	if err != nil {
		return err
	}

	return nil // OK
}

// ScanStart starts the scan operation
func (w *DBusService) ScanStart() *dbus.Error {
	if !w.connected {
		return newDBusError("HCI is disconnected")
	}

	if w.connecting {
		return newDBusError("Unable to scan, connection is in progress")
	}

	// Just let them know devices that are cached,
	// as they might be connected and no longer advertising.
	go func() {
		w.devicesLock.Lock()
		defer w.devicesLock.Unlock()
		rssi := 0 // RSSI is unknown

		for mac, pdev := range w.devicesDiscovered {
			w.emitPeripheralDiscovered(mac, pdev.name, rssi)
		}
	}()

	log.Infof("[%s] start scanning", TAG)
	w.device.Scan(nil, false)

	return nil // OK
}

// ScanStop stops the scan operation
func (w *DBusService) ScanStop() *dbus.Error {
	if !w.connected {
		return newDBusError("HCI is disconnected")
	}

	log.Infof("[%s] stop scanning", TAG)
	w.device.StopScanning()

	return nil // OK
}

// Connect connects to a remote device
func (w *DBusService) Connect(mac string, random bool) (bool, *dbus.Error) {
	mac, err := normalizeHex(mac)
	if err != nil {
		return false, newDBusError("Invalid MAC address provided")
	}

	log.WithField("mac", mac).WithField("random", random).
		Infof("[%s]: connecting", TAG)

	w.connecting = true
	defer func() {
		w.connecting = false
	}()

	var pdev *DeviceInfo
	if pdev = w.findDeviceDiscovered(mac); pdev == nil {
		log.WithField("mac", mac).Debugf("[%s]: no device discovered, try to get...", TAG)

		// no device discovered, try to get it
		b, _ := hex.DecodeString(mac)
		p, err := w.device.GetPeripheral(b, random)
		if err != nil {
			return false, newDBusError(fmt.Sprintf("Failed to get peripheral: %s", err))
		}

		// TODO: just call w.onPeripheralDiscovered(p, nil, 0)

		pdev = &DeviceInfo{
			name:       strings.Trim(p.Name(), "\x00"),
			peripheral: p,
		}

		w.insertDeviceDiscovered(mac, pdev)
		w.emitPeripheralDiscovered(mac, pdev.name, 0)
	}

	if connected := w.findDeviceConnected(mac); connected == nil {
		log.WithField("mac", mac).Debugf("[%s]: not connected yet, connecting now...", TAG)
		w.device.StopScanning()
		w.device.Connect(pdev.peripheral)

		select {
		case <-w.connChan:
			if !pdev.connectedOnce {
				pdev.characteristics = make(map[string]*gatt.Characteristic)

				done := make(chan bool, 1)

				go func() {
					pdev.explorePeripheral(pdev.peripheral)
					done <- true
				}()

				select {
				case <-done:
				case <-time.After(ExploreTimeout):
					return false, newDBusError("Peripheral explore timed out")
				}
			}

			pdev.connectedOnce = true
			pdev.ready = true

			log.WithField("mac", mac).Infof("[%s]: peripheral connected", TAG)
			w.emitPeripheralConnected(mac)

		case <-time.After(ConnectTimeout):
			w.device.CancelConnection(pdev.peripheral)
			return false, newDBusError("Connect timed out")
		}
	}

	return true, nil // OK
}

// Disconnect removes the peripheral
func (w *DBusService) Disconnect(mac string) *dbus.Error {
	mac, _ = normalizeHex(mac)

	if pdev := w.findDeviceDiscovered(mac); pdev != nil {
		// TODO: remove from w.devicesConnected?
		w.device.CancelConnection(pdev.peripheral)
		return nil // OK
	}

	return newDBusError("Not connected")
}

// Is peripheral connected?
func (w *DBusService) Connected(mac string) (bool, *dbus.Error) {
	mac, _ = normalizeHex(mac)
	pdev := w.findDeviceConnected(mac)
	return pdev != nil, nil
}

// GattWrite writes the GATT characteristic
func (w *DBusService) GattWrite(mac string, uuid string, message string) *dbus.Error {
	_, err := w.handleGattCommand(mac, uuid, message,
		func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
			return nil, p.WriteCharacteristic(c, b, false)
		})
	return err
}

// GattWriteNoResp writes the GATT characteristic without response
func (w *DBusService) GattWriteNoResp(mac string, uuid string, message string) *dbus.Error {
	_, err := w.handleGattCommand(mac, uuid, message,
		func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
			return nil, p.WriteCharacteristic(c, b, true)
		})
	return err
}

// GattRead reads the GATT characteristic
func (w *DBusService) GattRead(mac string, uuid string) (string, *dbus.Error) {
	message, err := w.handleGattCommand(mac, uuid, "",
		func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
			return p.ReadCharacteristic(c)
		})
	return message, err
}

// GattNotifications enables/disables GATT characteristic notifications
func (w *DBusService) GattNotifications(mac string, uuid string, enable bool) *dbus.Error {
	_, err := w.handleGattCommand(mac, uuid, "",
		func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
			if enable {
				return nil, p.SetNotifyValue(c,
					func(c *gatt.Characteristic, b []byte, e error) {
						if e != nil {
							log.WithError(e).Warnf("[%s]: notification handler received error", TAG)
							return
						}

						message := hex.EncodeToString(b)
						w.emitNotificationReceived(mac, uuid, message)
					})
			} else {
				return nil, p.SetNotifyValue(c, nil)
			}
		})

	return err
}

// GattIndications enables/disables GATT characteristic indications
func (w *DBusService) GattIndications(mac string, uuid string, enable bool) *dbus.Error {
	_, err := w.handleGattCommand(mac, uuid, "",
		func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
			if enable {
				return nil, p.SetIndicateValue(c,
					func(c *gatt.Characteristic, b []byte, e error) {
						if e != nil {
							log.WithError(e).Warnf("[%s]: indications handler received error", TAG)
							return
						}

						message := hex.EncodeToString(b)
						w.emitIndicationReceived(mac, uuid, message)
					})
			} else {
				return nil, p.SetIndicateValue(c, nil)
			}
		})

	return err
}

// "init state changed" handler
func (w *DBusService) onInit(dev gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		log.Infof("[%s]: HCI device is connected", TAG)
		w.connected = true
	default:
		log.WithField("state", s).Infof("[%s]: HCI device is disconnected", TAG)
		w.connected = false
	}
}

// "PeripheralDiscovered" handler
func (w *DBusService) onPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	mac, _ := normalizeHex(p.ID())
	name := strings.Trim(p.Name(), "\x00")

	if pdev := w.findDeviceDiscovered(mac); pdev != nil {
		// update existing peripheral
		if pdev.name == "" && name != "" {
			pdev.name = name
		}
		w.emitPeripheralDiscovered(mac, pdev.name, rssi)
	} else {
		// new peripheral discovered
		pdev = &DeviceInfo{
			name:       name,
			rssi:       rssi,
			peripheral: p,
		}
		w.insertDeviceDiscovered(mac, pdev)
		w.emitPeripheralDiscovered(mac, name, rssi)
	}
}

// "PeripheralConnected" handler
func (w *DBusService) onPeripheralConnected(p gatt.Peripheral, err error) {
	// TODO: check `err` error?
	mac, _ := normalizeHex(p.ID())

	if pdev := w.findDeviceDiscovered(mac); pdev != nil {
		pdev.peripheral = p

		log.WithField("mac", mac).Infof("[%s]: connected", TAG)
		w.insertDeviceConnected(mac, pdev)
		w.connChan <- true
	}
}

// "PeripheralDisconnected" handler
func (w *DBusService) onPeripheralDisconnected(p gatt.Peripheral, err error) {
	// TODO: check `err` error?
	mac, _ := normalizeHex(p.ID())

	log.WithField("mac", mac).Infof("[%s]: disconnected", TAG)

	w.removeDeviceDiscovered(mac)
	if w.removeDeviceConnected(mac) {
		w.emitPeripheralDisconnected(mac)
	}
}

// emit "PeripheralDiscovered" signal
func (w *DBusService) emitPeripheralDiscovered(ID, name string, rssi int) error {
	log.WithField("id", ID).WithField("name", name).WithField("rssi", rssi).
		Debugf("[%s]: emitting PeripheralDiscovered signal...", TAG)

	return w.bus.Emit(ComDevicehiveBluetoothPath,
		ComDevicehiveBluetoothIface+".PeripheralDiscovered",
		ID, name, int16(rssi))
}

// emit "PeripheralDisconnected" signal
func (w *DBusService) emitPeripheralDisconnected(ID string) error {
	log.WithField("id", ID).Debugf("[%s]: emitting PeripheralDisconnected signal...", TAG)

	return w.bus.Emit(ComDevicehiveBluetoothPath,
		ComDevicehiveBluetoothIface+".PeripheralDisconnected", ID)
}

// emit "PeripheralConnected" signal
func (w *DBusService) emitPeripheralConnected(ID string) error {
	log.WithField("id", ID).Debugf("[%s]: emitting PeripheralConnected signal...", TAG)

	return w.bus.Emit(ComDevicehiveBluetoothPath,
		ComDevicehiveBluetoothIface+".PeripheralConnected", ID)
}

// emit "NotificationReceived" signal
func (w *DBusService) emitNotificationReceived(mac, uuid, m string) error {
	log.WithField("MAC", mac).WithField("UUID", uuid).WithField("msg", m).
		Debugf("[%s]: emitting NotificationReceived signal...", TAG)

	return w.bus.Emit(ComDevicehiveBluetoothPath,
		ComDevicehiveBluetoothIface+".NotificationReceived",
		mac, uuid, m)
}

// emit "IndicationReceived" signal
func (w *DBusService) emitIndicationReceived(mac, uuid, m string) error {
	log.WithField("MAC", mac).WithField("UUID", uuid).WithField("msg", m).
		Debugf("[%s]: emitting IndicationReceived signal...", TAG)

	return w.bus.Emit(ComDevicehiveBluetoothPath,
		ComDevicehiveBluetoothIface+".IndicationReceived",
		mac, uuid, m)
}

// handle any GATT request
func (w *DBusService) handleGattCommand(mac string, uuid string, message string,
	handler func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error)) (string, *dbus.Error) {
	mac, _ = normalizeHex(mac)
	uuid, _ = normalizeHex(uuid)

	var res string
	if pdev := w.devicesDiscovered[mac]; pdev != nil {
		if !pdev.ready {
			log.WithField("mac", mac).Warnf("[%s]: peripheral is not ready yet", TAG)
			return "", newDBusError(fmt.Sprintf("Peripheral [%s] is not ready yet", mac))
		}

		if connected := w.findDeviceConnected(mac); connected == nil {
			log.WithField("mac", mac).Warnf("[%s]: peripheral is not connected", TAG)
			return "", newDBusError(fmt.Sprintf("Peripheral [%s] is not connected", mac))
		}

		pdev.lock.Lock()
		defer pdev.lock.Unlock()

		log.WithField("mac", mac).WithField("char", uuid).WithField("message", message).
			Infof("[%s]: sending GATT command...", TAG)

		var bmsg []byte
		var bres []byte
		var err error

		if message != "" {
			bmsg, err = hex.DecodeString(message)
			if err != nil {
				log.WithError(err).WithField("message", message).
					Warnf("[%s]: failed to decode message from HEX", TAG)
				return "", newDBusError(fmt.Sprintf("Failed to decode message from HEX: %s", err))
			}
		}

		if c, ok := pdev.characteristics[uuid]; ok {
			bres, err = handler(pdev.peripheral, c, bmsg)
			if bres != nil {
				res = hex.EncodeToString(bmsg)
			}

			if err != nil {
				log.WithError(err).Warnf("[%s]: failed to write or read characteristic", TAG)
				return "", newDBusError(fmt.Sprintf("Failed to write/read characteristic: %s", err))
			}
		} else {
			log.WithField("char", uuid).Warnf("[%s]: characteristic not found", TAG)
			return "", newDBusError(fmt.Sprintf("Characteristic [%s] not found", uuid))
		}
	} else {
		log.WithField("mac", mac).Warnf("[%s]: peripheral is not discovered", TAG)
		return "", newDBusError(fmt.Sprintf("Peripheral [%s] is not discovered", mac))
	}

	return res, nil // OK
}

// get existing peripheral discovered, nil if not found
func (w *DBusService) findDeviceDiscovered(mac string) *DeviceInfo {
	w.devicesLock.Lock()
	defer w.devicesLock.Unlock()

	if pdev, ok := w.devicesDiscovered[mac]; ok {
		return pdev
	}

	return nil // not found
}

// insert new peripheral discovered, existing item will be replaced
func (w *DBusService) insertDeviceDiscovered(mac string, pdev *DeviceInfo) bool {
	w.devicesLock.Lock()
	defer w.devicesLock.Unlock()

	if pdev != nil {
		w.devicesDiscovered[mac] = pdev
		return true // updated
	}

	return false
}

// remove peripheral discovered
func (w *DBusService) removeDeviceDiscovered(mac string) bool {
	w.devicesLock.Lock()
	defer w.devicesLock.Unlock()

	if _, ok := w.devicesDiscovered[mac]; ok {
		delete(w.devicesDiscovered, mac)
		return true
	}

	return false // not found
}

// get existing peripheral connected, nil if not found
func (w *DBusService) findDeviceConnected(mac string) *DeviceInfo {
	w.devicesLock.Lock()
	defer w.devicesLock.Unlock()

	if pdev, ok := w.devicesConnected[mac]; ok {
		return pdev
	}

	return nil // not found
}

// insert new peripheral connected, existing item will be replaced
func (w *DBusService) insertDeviceConnected(mac string, pdev *DeviceInfo) bool {
	w.devicesLock.Lock()
	defer w.devicesLock.Unlock()

	if pdev != nil {
		w.devicesConnected[mac] = pdev
		return true
	}

	return false
}

// remove peripheral connected
func (w *DBusService) removeDeviceConnected(mac string) bool {
	w.devicesLock.Lock()
	defer w.devicesLock.Unlock()

	if _, ok := w.devicesConnected[mac]; ok {
		delete(w.devicesConnected, mac)
		return true
	}

	return false // not found
}

// explore peripheral
func (pdev *DeviceInfo) explorePeripheral(p gatt.Peripheral) error {
	// discover all services
	services, err := p.DiscoverServices(nil)
	if err != nil {
		log.WithError(err).Warnf("[%s] failed to discover services", TAG)
		return fmt.Errorf("failed to discover services: %s", err)
	}

	// check each service
	for _, s := range services {
		log.WithField("uuid", s.UUID().String()).WithField("name", s.Name()).
			Debugf("[%s]: new service discovered", TAG)

		// discover characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			log.WithError(err).Warnf("[%s]: failed to discover characteristics, ignored", TAG)
			continue // ignore error
		}

		// check each characteristic
		for _, c := range cs {
			// msg := "  Characteristic  " + c.UUID().String()
			// if len(c.Name()) > 0 {
			// 	msg += " (" + c.Name() + ")"
			// }
			// msg += "\n    properties    " + c.Properties().String()
			// log.Println(msg)
			log.WithField("uuid", c.UUID().String()).WithField("name", c.Name()).
				WithField("properties", c.Properties().String()).
				Debugf("[%s]: new characteristic discovered", TAG)

			// discover descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				log.WithError(err).Warnf("[%s]: failed to discover descriptors, ignored", TAG)
				continue // ignore error
			}

			// check each descriptor
			for _, d := range ds {
				log.WithField("uuid", d.UUID().String()).WithField("name", d.Name()).
					Debugf("[%s]: new descriptor discovered", TAG)
			}

			id, _ := normalizeHex(c.UUID().String())
			pdev.characteristics[id] = c
		}
	}

	pdev.ready = true
	return nil
}

// create new DBus error
func newDBusError(message string) *dbus.Error {
	return dbus.NewError("com.devicehive.Error",
		[]interface{}{message})
}

// this functions removes all ":", "-", " " symbols
func normalizeHex(s string) (res string, err error) {
	trimmed := strings.Map(func(r rune) rune {
		if !strings.ContainsRune(":- ", r) {
			return r
		}
		return -1
	}, s)

	b, err := hex.DecodeString(trimmed)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b), nil
}
