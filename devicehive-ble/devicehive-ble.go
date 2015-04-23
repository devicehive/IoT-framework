package main

import (
	"encoding/hex"
	"fmt"
	"github.com/godbus/dbus"
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

func newDHError(message string) *dbus.Error {
	return dbus.NewError("com.devicehive.Error",
		[]interface{}{message})
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
		if _, ok := wrapper.devicesDiscovered[id]; !ok {
			wrapper.devicesDiscovered[id] = &DiscoveredDeviceInfo{name: p.Name(), rssi: rssi, peripheral: p}
			log.Printf("Adding mac: %s", id)
		}
		bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.DeviceDiscovered", id, p.Name(), int16(rssi))
	}))

	d.Handle(gatt.PeripheralConnected(func(p gatt.Peripheral, err error) {
		id, _ := normalizeHex(p.ID())
		if _, ok := wrapper.devicesDiscovered[id]; ok {
			dev := wrapper.devicesDiscovered[id]
			dev.characteristics = make(map[string]*gatt.Characteristic)
			dev.ready = false
			dev.peripheral = p
			dev.explorePeripheral(dev.peripheral)
			dev.ready = true
			bus.Emit("/com/devicehive/bluetooth", "com.devicehive.bluetooth.DeviceConnected", id)
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

func (w *BleDbusWrapper) ScanStart() (string, *dbus.Error) {
	if !w.connected {
		return "ERROR", newDHError("Disconnected")
	}

	w.device.Scan(nil, false)
	return "OK", nil
}

func (w *BleDbusWrapper) ScanStop() (string, *dbus.Error) {
	if !w.connected {
		return "ERROR", newDHError("Disconnected")
	}

	w.device.StopScanning()
	return "OK", nil
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

func (w *BleDbusWrapper) Connect(mac string) (string, *dbus.Error) {
	mac, err := normalizeHex(mac)

	if err != nil {
		return "ERROR", newDHError("Invalid MAC provided")
	}

	log.Printf("Connecting to: %s", mac)

	if val, ok := w.devicesDiscovered[mac]; ok {
		w.device.Connect(val.peripheral)
		return "OK", nil
	}

	log.Print("Connect(): MAC wasn't descovered, use Scan Start/Stop before")
	return "Undiscovered MAC", newDHError("Connect(): MAC wasn't descovered, use Scan Start/Stop before")
}

func (w *BleDbusWrapper) Disconnect(mac string) (string, *dbus.Error) {
	if val, ok := w.devicesDiscovered[mac]; ok {
		w.device.CancelConnection(val.peripheral)
		return "OK", nil
	}
	
	log.Print("Disconnect(): MAC wasn't descovered, use Scan Start/Stop before")
	return "ERROR", newDHError("Disconnect(): MAC wasn't descovered, use Scan Start/Stop before")
}

func (w *BleDbusWrapper) GattWrite(mac string, uuid string, message string) (string, *dbus.Error) {
	mac, _ = normalizeHex(mac)
	uuid, _ = normalizeHex(uuid)

	b, err := hex.DecodeString(message)

	if err != nil {
		log.Printf("Invalid message: %s", message)
		return "ERROR", newDHError(fmt.Sprintf("Invalid message: %s", message))
	}

	if val, ok := w.devicesDiscovered[mac]; ok {
		log.Printf("Writing: %v to mac %v char %v", b, mac, uuid)
		error := val.peripheral.WriteCharacteristic(val.characteristics[uuid], b, false)

		if error != nil {
			log.Printf("Error writing characteristic: %s", error)
			return "ERROR", newDHError(error.Error())
		}

		return "OK", nil
	}

	log.Printf("Invalid peripheral ID: %s", mac)
	return "ERROR", newDHError("Invalid peripheral ID")
}

func (b *DiscoveredDeviceInfo) explorePeripheral(p gatt.Peripheral) {
	fmt.Println("Connected")
	//defer p.Device().CancelConnection(p)

	// Discovery services
	ss, err := p.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services, err: %s\n", err)
		return
	}

	for _, s := range ss {
		msg := "Service: " + s.UUID().String()
		if len(s.Name()) > 0 {
			msg += " (" + s.Name() + ")"
		}
		fmt.Println(msg)

		// Discovery characteristics
		cs, err := p.DiscoverCharacteristics(nil, s)
		if err != nil {
			fmt.Printf("Failed to discover characteristics, err: %s\n", err)
			continue
		}

		for _, c := range cs {
			msg := "  Characteristic  " + c.UUID().String()
			if len(c.Name()) > 0 {
				msg += " (" + c.Name() + ")"
			}
			msg += "\n    properties    " + c.Properties().String()
			fmt.Println(msg)

			// Read the characteristic, if possible.
			if (c.Properties() & gatt.CharRead) != 0 {
				b, err := p.ReadCharacteristic(c)
				if err != nil {
					fmt.Printf("Failed to read characteristic, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			// Discovery descriptors
			ds, err := p.DiscoverDescriptors(nil, c)
			if err != nil {
				fmt.Printf("Failed to discover descriptors, err: %s\n", err)
				continue
			}

			for _, d := range ds {
				msg := "  Descriptor      " + d.UUID().String()
				if len(d.Name()) > 0 {
					msg += " (" + d.Name() + ")"
				}
				fmt.Println(msg)

				// Read descriptor (could fail, if it's not readable)
				b, err := p.ReadDescriptor(d)
				if err != nil {
					fmt.Printf("Failed to read descriptor, err: %s\n", err)
					continue
				}
				fmt.Printf("    value         %x | %q\n", b, b)
			}

			id, _ := normalizeHex(c.UUID().String())
			b.characteristics[id] = c

		}
		fmt.Println()
		b.ready = true
	}
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

	log.Printf("Exporting BleDbusWrapper...")
	bus.Export(w, "/com/devicehive/bluetooth", "com.devicehive.bluetooth")

	// d, err := gatt.NewDevice([]gatt.Option{
	// 	gatt.LnxDeviceID(0, false),
	// }...)

	// var device gatt.Peripheral

	// if err != nil {
	// 	log.Fatalf("Failed to open device, err: %s\n", err)
	// 	return
	// }

	// // Register handlers.
	// d.Handle(gatt.PeripheralDiscovered(func(p gatt.Peripheral, a *gatt.Advertisement, rssi int) {
	// 	fmt.Printf("\nPeripheral ID:%s, NAME:(%s)\n", p.ID(), p.Name())
	// 	if p.Name() == "Weight Measurement" {
	// 		d.Connect(p)
	// 		d.StopScanning()
	// 		device = p
	// 	}
	// }))

	// dev := NewBleDevice()

	// d.Handle(gatt.PeripheralConnected(dev.onPeriphConnected))
	// d.Init(onStateChanged)

	select {}
}

