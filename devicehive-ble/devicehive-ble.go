package main

// import _ "net/http/pprof"

import (
	"encoding/hex"
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"
	"github.com/devicehive/gatt"
	"log"
	//	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ConnectionTimeout = 3 // Connection timeout in seconds.
	ExploreTimeout    = 5 // Timeout to explore peripherals for a newly found device
)

type BleDbusWrapper struct {
	bus                   *dbus.Conn
	device                gatt.Device
	connected             bool
	devicesDiscovered     map[string]*DiscoveredDeviceInfo
	devicesDiscoveredsync sync.Mutex
	devicesConnected      map[string]*sync.Mutex
	devicesConnectedsync  sync.Mutex
	connChan              chan bool
	connecting            bool
}

type DiscoveredDeviceInfo struct {
	name            string
	rssi            int
	peripheral      gatt.Peripheral
	characteristics map[string]*gatt.Characteristic
	ready           bool
	connectedOnce   bool
}

type gattCommandHandler func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error)

func newDHError(message string) *dbus.Error {
	return dbus.NewError("com.devicehive.Error",
		[]interface{}{message})
}

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

func NewBleDbusWrapper(bus *dbus.Conn) *BleDbusWrapper {
	d, err := gatt.NewDevice([]gatt.Option{
		gatt.LnxDeviceID(0, false),
	}...)

	if err != nil {
		log.Fatalf("Failed to open device, err: %s\n", err)
		return nil
	}

	wrapper := new(BleDbusWrapper)
	wrapper.bus = bus
	wrapper.devicesDiscovered = make(map[string]*DiscoveredDeviceInfo)
	wrapper.devicesConnected = make(map[string]*sync.Mutex)
	wrapper.connChan = make(chan bool, 1)

	d.Handle(gatt.PeripheralDiscovered(wrapper.OnPeripheralDiscovered))
	d.Handle(gatt.PeripheralConnected(wrapper.OnPeripheralConnected))
	d.Handle(gatt.PeripheralDisconnected(wrapper.OnPeripheralDisconnected))
	d.Init(wrapper.OnInit)

	wrapper.device = d

	return wrapper
}

func (w *BleDbusWrapper) OnInit(dev gatt.Device, s gatt.State) {
	switch s {
	case gatt.StatePoweredOn:
		log.Print("HCI device powered on")
		w.connected = true
	default:
		log.Printf("StateChanged handler received: %v", s)
		w.connected = false
	}
}

func (w *BleDbusWrapper) OnPeripheralConnected(p gatt.Peripheral, err error) {
	id, _ := normalizeHex(p.ID())
	w.devicesDiscovered[id].peripheral = p

	w.devicesConnectedsync.Lock()
	w.devicesConnected[id] = &sync.Mutex{}
	w.devicesConnectedsync.Unlock()

	w.connChan <- true
}

func (w *BleDbusWrapper) OnPeripheralDisconnected(p gatt.Peripheral, err error) {
	w.devicesConnectedsync.Lock()
	defer w.devicesConnectedsync.Unlock()

	id, _ := normalizeHex(p.ID())

	if _, ok := w.devicesConnected[id]; ok {
		log.Printf("Disconnected: %s", id)
		delete(w.devicesConnected, id)
		w.emitPeripheralDisconnected(id)
	}
}

func (w *BleDbusWrapper) OnPeripheralDiscovered(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	w.devicesDiscoveredsync.Lock()
	defer w.devicesDiscoveredsync.Unlock()

	id, _ := normalizeHex(p.ID())
	name := strings.Trim(p.Name(), "\x00")

	dev, ok := w.devicesDiscovered[id]
	if !ok {
		w.devicesDiscovered[id] = &DiscoveredDeviceInfo{name: name, rssi: rssi, peripheral: p, ready: false, connectedOnce: false}
		w.emitPeripheralDiscovered(id, name, int16(rssi))
	} else {
		if (dev.name == "") && (name != "") {
			dev.name = name
			w.emitPeripheralDiscovered(id, name, int16(rssi))
		}
	}
}

func (w *BleDbusWrapper) emitPeripheralDiscovered(id, name string, rssi int16) {
	w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.PeripheralDiscovered", id, name, int16(rssi))
}

func (w *BleDbusWrapper) emitPeripheralDisconnected(id string) {
	w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.PeripheralDisconnected", id)
}

func (w *BleDbusWrapper) emitPeripheralConnected(id string) {
	w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.PeripheralConnected", id)
}

func (w *BleDbusWrapper) emitNotificationReceived(mac, uuid, m string) {
	w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.NotificationReceived", mac, uuid, m)
}

func (w *BleDbusWrapper) emitIndicationReceived(mac, uuid, m string) {
	w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.IndicationReceived", mac, uuid, m)
}

func (w *BleDbusWrapper) ScanStart() *dbus.Error {
	if !w.connected {
		return newDHError("HCI is disconnected")
	}

	// Just let them know devices that are cached, as they might be connected and
	// no longer advertising. Put RSSI to 0.
	go func() {
		for k, v := range w.devicesDiscovered {
			w.emitPeripheralDiscovered(k, v.name, int16(0))
		}
	}()

	if w.connecting {
		return newDHError("Unable to scan, connection is in progress")
	}

	w.device.Scan(nil, false)

	return nil
}

func (w *BleDbusWrapper) ScanStop() *dbus.Error {
	if !w.connected {
		return newDHError("HCI is disconnected")
	}

	w.device.StopScanning()
	return nil
}

func (w *BleDbusWrapper) Connect(mac string, random bool) (bool, *dbus.Error) {
	mac, err := normalizeHex(mac)
	w.connecting = true
	defer func() { w.connecting = false }()

	if err != nil {
		return false, newDHError("Invalid MAC provided")
	}

	val, ok := w.devicesDiscovered[mac]

	if !ok {
		b, _ := hex.DecodeString(mac)
		p, err := w.device.GetPeripheral(b, random)

		if err != nil {
			return false, newDHError(err.Error())
		}

		val = &DiscoveredDeviceInfo{name: "", rssi: 0, peripheral: p, ready: false, connectedOnce: false}
		w.devicesDiscoveredsync.Lock()
		w.devicesDiscovered[mac] = val
		w.devicesDiscoveredsync.Unlock()
	}

	if _, ok := w.devicesConnected[mac]; !ok {
		log.Printf("Trying to connect: %s", mac)
		w.device.StopScanning()
		w.device.Connect(val.peripheral)
		select {
		case <-w.connChan:
			id, _ := normalizeHex(mac)
			log.Printf("PeripheralConnected: %s", id)
			if !val.connectedOnce {
				val.characteristics = make(map[string]*gatt.Characteristic)

				done := make(chan bool, 1)

				go func() {
					val.explorePeripheral(val.peripheral)
					done <- true
				}()

				select {
				case <-done:
				case <-time.After(ExploreTimeout * time.Second):
					return false, newDHError("BLE peripheral explore timeout")
				}

			}
			val.connectedOnce = true
			val.ready = true
			w.emitPeripheralConnected(mac)
			log.Printf("Connected to: %s", mac)

		case <-time.After(ConnectionTimeout * time.Second):
			w.device.CancelConnection(val.peripheral)
			return false, newDHError(fmt.Sprintf("BLE connection timed out [%s]", mac))
		}
	}

	return true, nil
}

func (w *BleDbusWrapper) Disconnect(mac string) *dbus.Error {
	if val, ok := w.devicesDiscovered[mac]; ok {
		w.device.CancelConnection(val.peripheral)
		return nil
	}

	return newDHError(fmt.Sprintf("MAC [%s] wasn't descovered, use Scan Start/Stop first", mac))
}

func (w *BleDbusWrapper) handleGattCommand(mac string, uuid string, message string, handler gattCommandHandler) (string, *dbus.Error) {
	mac, _ = normalizeHex(mac)
	uuid, _ = normalizeHex(uuid)

	deviceLock, ok := w.devicesConnected[mac]
	if ok && deviceLock != nil {
		deviceLock.Lock()
		defer deviceLock.Unlock()
	} else {
		return "", newDHError(fmt.Sprintf("Device [%s] not connected", mac))
	}

	res := ""

	if val, ok := w.devicesDiscovered[mac]; ok {
		if !val.ready {
			log.Printf("Device %s is not ready (probably still connecting, or discoverig services and characteristics)", mac)
			return "", newDHError("Device not ready")
		}

		if _, ok := w.devicesConnected[mac]; !ok {
			log.Printf("handleGattCommand(): %s is not connected", mac)
			return "", newDHError(fmt.Sprintf("Device [%s] not connected", mac))
		}

		log.Printf("Sending gatt command [mac: %s, char %s, message: %s]", mac, uuid, message)
		var b []byte
		var err error

		if message != "" {
			b, err = hex.DecodeString(message)
			if err != nil {
				log.Printf("Invalid message: %s", message)
				return "", newDHError(err.Error())
			}
		}

		if c, ok := val.characteristics[uuid]; ok {
			b, err = handler(val.peripheral, c, b)

			if b != nil {
				res = hex.EncodeToString(b)
			}

			if err != nil {
				log.Printf("Error writing/reading characteristic: %s", err)
				return "", newDHError(err.Error())
			}
		} else {
			s := fmt.Sprintf("Characteristic %s not found. Please try full name and check the device spec.", uuid)
			log.Print(s)
			return "", newDHError(s)
		}

	} else {
		log.Printf("Invalid peripheral ID: %s", mac)
		return "", newDHError("Invalid peripheral ID")
	}

	return res, nil
}

func (w *BleDbusWrapper) GattWrite(mac string, uuid string, message string) *dbus.Error {
	h := func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
		error := p.WriteCharacteristic(c, b, false)
		return nil, error
	}

	_, error := w.handleGattCommand(mac, uuid, message, h)
	return error
}

func (w *BleDbusWrapper) GattWriteNoResp(mac string, uuid string, message string) *dbus.Error {
	h := func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
		error := p.WriteCharacteristic(c, b, true)
		return nil, error
	}

	_, error := w.handleGattCommand(mac, uuid, message, h)
	return error
}

func (w *BleDbusWrapper) Connected(mac string) (bool, *dbus.Error) {
	mac, _ = normalizeHex(mac)

	_, ok := w.devicesConnected[mac]

	return ok, nil
}

func (w *BleDbusWrapper) GattRead(mac string, uuid string) (string, *dbus.Error) {
	h := func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
		return p.ReadCharacteristic(c)
	}

	return w.handleGattCommand(mac, uuid, "", h)
}

func (w *BleDbusWrapper) GattNotifications(mac string, uuid string, enable bool) *dbus.Error {
	h := func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {

		if enable {
			err := p.SetNotifyValue(c, func(_ *gatt.Characteristic, b []byte, e error) {
				if e != nil {
					log.Printf("Notification handler received error: %s", e)
					return
				}

				m := hex.EncodeToString(b)
				w.emitNotificationReceived(mac, uuid, m)
			})
			return nil, err
		} else {
			return nil, p.SetNotifyValue(c, nil)
		}
	}

	_, err := w.handleGattCommand(mac, uuid, "", h)
	return err
}

func (w *BleDbusWrapper) GattIndications(mac string, uuid string, enable bool) *dbus.Error {
	h := func(p gatt.Peripheral, c *gatt.Characteristic, b []byte) ([]byte, error) {
		if enable {
			err := p.SetIndicateValue(c, func(_ *gatt.Characteristic, b []byte, e error) {
				if e != nil {
					log.Printf("Indications handler received error: %s", e)
					return
				}

				m := hex.EncodeToString(b)
				w.emitIndicationReceived(mac, uuid, m)
			})
			return nil, err
		} else {
			return nil, p.SetIndicateValue(c, nil)
		}
	}

	_, err := w.handleGattCommand(mac, uuid, "", h)
	return err
}

func (b *DiscoveredDeviceInfo) explorePeripheral(p gatt.Peripheral) error {
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return err
	}

	for _, s := range ss {
		// msg := "Service: " + s.UUID().String()
		// if len(s.Name()) > 0 {
		// 	msg += " (" + s.Name() + ")"
		// }
		// log.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			log.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			// msg := "  Characteristic  " + c.UUID().String()
			// if len(c.Name()) > 0 {
			// 	msg += " (" + c.Name() + ")"
			// }
			// msg += "\n    properties    " + c.Properties().String()
			// log.Println(msg)

			// Discovery descriptors
			_, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				log.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			// for _, d := range ds {
			// 	msg := "  Descriptor      " + d.UUID().String()
			// 	if len(d.Name()) > 0 {
			// 		msg += " (" + d.Name() + ")"
			// 	}
			// 	log.Println(msg)
			// }

			id, _ := normalizeHex(c.UUID().String())
			b.characteristics[id] = c
		}
		b.ready = true
	}

	return nil
}

func main() {
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	var err error
	var bus *dbus.Conn
	bus, err = dbus.SystemBus()
	if err != nil {
		log.Panic(err)
	}

	reply, err := bus.RequestName("com.devicehive.bluetooth",
		dbus.NameFlagDoNotQueue)
	if err != nil {
		log.Panic(err)
	}

	if reply != dbus.RequestNameReplyPrimaryOwner {
		log.Fatal("name already taken")
	}

	w := NewBleDbusWrapper(bus)
	bus.Export(w, "/com/devicehive/bluetooth", "com.devicehive.bluetooth")

	// Introspectable
	n := &introspect.Node{
		Name: "/com/devicehive/bluetooth",
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:    "com.devicehive.bluetooth",
				Methods: introspect.Methods(w),
			},
		},
	}

	bus.Export(introspect.NewIntrospectable(n), "/com/devicehive/bluetooth", "org.freedesktop.DBus.Introspectable")

	select {}
}
