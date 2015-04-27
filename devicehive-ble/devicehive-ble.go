package main

import (
	"encoding/hex"
	"fmt"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/godbus/dbus/prop"
	"github.com/paypal/gatt"
	"log"
	"strings"
)

type BleDbusWrapper struct {
	bus               *dbus.Conn
	device            gatt.Device
	connected         bool
	devicesDiscovered map[string]*DiscoveredDeviceInfo
}

type DiscoveredDeviceInfo struct {
	name            string
	rssi            int
	peripheral      gatt.Peripheral
	characteristics map[string]*gatt.Characteristic
	ready           bool
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
	wrapper.device = d
	wrapper.devicesDiscovered = make(map[string]*DiscoveredDeviceInfo)

	d.Handle(gatt.PeripheralDiscovered(func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
		id, _ := normalizeHex(p.ID())
		name := strings.Trim(p.Name(), "\x00")
		if _, ok := wrapper.devicesDiscovered[id]; !ok {
			wrapper.devicesDiscovered[id] = &DiscoveredDeviceInfo{name: name, rssi: rssi, peripheral: p, ready: false}
			log.Printf("Adding mac: %s - %s", id, name)
		}
		bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.DeviceDiscovered", id, name, int16(rssi))
	}))

	d.Handle(gatt.PeripheralConnected(func(p gatt.Peripheral, err error) {
		id, _ := normalizeHex(p.ID())
		if _, ok := wrapper.devicesDiscovered[id]; ok {
			dev := wrapper.devicesDiscovered[id]
			dev.characteristics = make(map[string]*gatt.Characteristic)			
			dev.peripheral = p
			dev.explorePeripheral(dev.peripheral)
			dev.ready = true
			bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.DeviceConnected", id)
		}
	}))

	d.Handle(gatt.PeripheralDisconnected(func(p gatt.Peripheral, err error) {
		id, _ := normalizeHex(p.ID())
		if dev, ok := wrapper.devicesDiscovered[id]; ok {
			dev.ready = false
			bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.DeviceDisconnected", id)
		}
	}))

	d.Init(func(dev gatt.Device, s gatt.State) {
		switch s {
		case gatt.StatePoweredOn:
			log.Print("HCI device powered on")
			wrapper.connected = true
			return
		default:
			log.Printf("StateChanged handler received: %v", s)
			wrapper.connected = false
		}
	})
	wrapper.device = d
	return wrapper
}

func (w *BleDbusWrapper) ScanStart() *dbus.Error {
	if !w.connected {
		return newDHError("Disconnected")
	}

	w.device.Scan(nil, false)

	// Just let them know devices that are cached, as they might be connected and
	// no longer advertising. Put RSSI to 0.
	go func() {
		for k, v := range w.devicesDiscovered {
			w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.DeviceDiscovered", k, v.name, int16(0))
		}
	}()

	return nil
}

func (w *BleDbusWrapper) ScanStop() *dbus.Error {
	if !w.connected {
		return newDHError("Disconnected")
	}

	w.device.StopScanning()
	return nil
}

func (w *BleDbusWrapper) Connect(mac string) (bool, *dbus.Error) {
	mac, err := normalizeHex(mac)

	if err != nil {
		return false, newDHError("Invalid MAC provided")
	}

	log.Printf("Connecting to: %s", mac)

	if val, ok := w.devicesDiscovered[mac]; ok {

		if !val.ready {
			log.Printf("trying to connect: %s", mac)
			w.device.Connect(val.peripheral)
		} else {
			log.Printf("Already connected to: %s", mac)
		}

		return val.ready, nil
	}

	log.Print("MAC wasn't descovered")
	return false, newDHError("MAC wasn't descovered, use Scan Start/Stop first")
}

func (w *BleDbusWrapper) Disconnect(mac string) *dbus.Error {
	if val, ok := w.devicesDiscovered[mac]; ok {
		w.device.CancelConnection(val.peripheral)
		return nil
	}

	log.Print("MAC wasn't descovered")
	return newDHError("MAC wasn't descovered, use Scan Start/Stop first")
}

func (w *BleDbusWrapper) handleGattCommand(mac string, uuid string, message string, handler gattCommandHandler) (string, *dbus.Error) {
	mac, _ = normalizeHex(mac)
	uuid, _ = normalizeHex(uuid)

	res := ""

	if val, ok := w.devicesDiscovered[mac]; ok {
		if !w.devicesDiscovered[mac].ready {
			log.Printf("Device %s is not ready (probably still connecting, or discoverig services and characteristics)", mac)
			return "", newDHError("Device not ready")
		}

		log.Printf("GATT COMMAND to mac %v char %v", mac, uuid)
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
				log.Printf("Received notification: %s", m)
				w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.NotificationReceived", mac, uuid, m)
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
				log.Printf("Received indication: %s", m)
				w.bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.IndicationReceived", mac, uuid, m)
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
	log.Println("Connected")

	ss, err := p.DiscoverServices(nil)
	if err != nil {
		log.Printf("Failed to discover services, err: %s\n", err)
		return err
	}

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		log.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			log.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			log.Println(msg)

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				log.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				log.Println(msg)
			}

			id, _ := normalizeHex(c.UUID().String())
			b.characteristics[id] = c
		}
		log.Println("Done exploring peripheral")
		b.ready = true
	}

	return nil
}

func main() {
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
	n := &introspect.Node {
		Name: "/com/devicehive/bluetooth",
		Interfaces: []introspect.Interface {
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:       "com.devicehive.bluetooth",
				Methods:    introspect.Methods(w),
			},
		},
	}

	bus.Export(introspect.NewIntrospectable(n), "/com/devicehive/bluetooth", "org.freedesktop.DBus.Introspectable")


	select {}
}
