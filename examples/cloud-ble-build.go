package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/godbus/dbus"
	"github.com/montanaflynn/stats"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type dbusWrapper struct {
	conn         *dbus.Conn
	path, iface  string
	handlers     map[string]signalHandlerFunc
	handlersSync sync.Mutex

	deviceMap map[string]*deviceInfo
	scanMap   map[string]bool

	readingsBuffer map[string]*[]float64

	deviceLock sync.Mutex
}

type deviceInfo struct {
	name      string
	connected bool
}

type signalHandlerFunc func(args ...interface{})
type cloudCommandHandler func(map[string]interface{}) (map[string]interface{}, error)

type melody struct {
	w *dbusWrapper
}

func NewMelody(w *dbusWrapper) *melody {
	m := &melody{w: w}

	return m
}

func (m *melody) Play(mac string, s string, bpm int) (err error) {
	notes := map[rune]string{'c': "060000",
		'd': "0603e8",
		'e': "0607d0",
		'f': "060bb8",
		'g': "060fa0",
		'a': "061388",
		'b': "061770",
		'C': "061b58"}

	noteLen := 1
	noteLenStr := ""
	prev := '\x00'
	m.w.BleGattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", "c804")
	time.Sleep(1 * time.Second)
	for _, c := range s {
		log.Printf("c = %d", c)
		note, ok := notes[c]
		switch {
		case '0' <= c && c <= '9':
			if '0' <= prev && prev <= '9' {
				noteLenStr = noteLenStr + string(c)
			} else {
				noteLenStr = string(c)
			}
			prev = c
			continue
		case (ok || (c == '*')):
			if '0' <= prev && prev <= '9' {
				l, _ := strconv.ParseInt(noteLenStr, 0, 32)
				noteLen = int(l)
				log.Printf("Note length: %d", noteLen)
			}
			if ok {
				m.w.BleGattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", "10")
				time.Sleep(100 * time.Millisecond)
				log.Printf("Playing %s", note)
				m.w.BleGattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", note)
			}
		}
		prev = c
		wait := ((60000 / bpm) * 4 / noteLen)
		log.Printf("Wait: %d", wait)
		time.Sleep(time.Duration(wait) * time.Millisecond)
	}
	m.w.BleGattWriteNoResp(mac, "af230002-879d-6186-1f49-deca0e85d9c1", "10")

	return nil
}

func NewdbusWrapper(path string, iface string) (*dbusWrapper, error) {
	d := new(dbusWrapper)

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Panic(err)
	}

	d.handlers = make(map[string]signalHandlerFunc)

	d.conn = conn
	d.path = path
	d.iface = iface
	d.readingsBuffer = make(map[string]*[]float64)
	d.scanMap = make(map[string]bool)

	filter := fmt.Sprintf("type='signal',path='%[1]s',interface='%[2]s',sender='%[2]s'", path, iface)
	log.Printf("Filter: %s", filter)

	conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").Call("org.freedesktop.DBus.AddMatch", 0, filter)

	go func() {
		ch := make(chan *dbus.Signal, 1)
		conn.Signal(ch)
		for signal := range ch {
			if !((strings.Index(signal.Name, iface) == 0) && (string(signal.Path) == path)) {
				continue
			}

			// Dirty hack! Should go away one we move throttling to DH cloud service
			if signal.Name == "com.devicehive.bluetooth.PeripheralDiscovered" {
				mac := strings.ToLower(signal.Body[0].(string))
				name := strings.ToLower(signal.Body[1].(string))

				if _, ok := d.scanMap[name]; ok {
					if _, ok := d.deviceMap[mac]; !ok {
						log.Printf("Discovered new device [mac: %s, name: %s]", mac, name)
						d.deviceLock.Lock()
						d.deviceMap[mac] = &deviceInfo{name: name, connected: false}
						d.deviceLock.Unlock()
					}
				}
			}

			if handler, ok := d.handlers[signal.Name]; ok {
				handler(signal.Body)
			} else {
				log.Printf("Unhandled signal: %s", signal.Name)
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

func (d *dbusWrapper) SendNotification(name string, parameters interface{}, priority uint64) {
	b, _ := json.Marshal(parameters)
	d.call("SendNotification", name, string(b), priority)
}

func (d *dbusWrapper) RegisterHandler(signal string, h signalHandlerFunc) {
	d.handlersSync.Lock()
	d.handlers[d.iface+"."+signal] = h
	d.handlersSync.Unlock()
}

func (d *dbusWrapper) BleScanStart() error {
	c := d.call("ScanStart")
	return c.Err
}

func (d *dbusWrapper) BleConnect(mac string, random bool) error {
	c := d.call("Connect", mac, random)
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

func (d *dbusWrapper) BleGattWriteNoResp(mac, uuid, message string) (map[string]interface{}, error) {
	c := d.call("GattWriteNoResp", mac, uuid, message)
	return nil, c.Err
}

func (d *dbusWrapper) Connected(mac string) (map[string]interface{}, error) {
	r := false
	err := d.call("Connected", mac).Store(&r)

	res := map[string]interface{}{
		"value": r,
	}

	return res, err
}

func (d *dbusWrapper) BleGattNotifications(mac, uuid string, enable bool) (map[string]interface{}, error) {
	c := d.call("GattNotifications", mac, uuid, enable)
	return nil, c.Err
}

func (d *dbusWrapper) BleGattIndications(mac, uuid string, enable bool) (map[string]interface{}, error) {
	c := d.call("GattIndications", mac, uuid, enable)
	return nil, c.Err
}

func (d *dbusWrapper) Init(mac, name string) error {
	d.deviceLock.Lock()
	defer d.deviceLock.Unlock()

	// If mac is blank, we'll put the name on scan list and will keep adding
	// new devices discovered with that name to the list of devices to connect to
	if mac == "" {
		n := strings.ToLower(name)
		log.Printf("Adding %s to scan list", name)
		if _, ok := d.scanMap[n]; !ok {
			d.scanMap[n] = false
		}
	} else {
		m := strings.ToLower(mac)
		if _, ok := d.deviceMap[m]; !ok {
			d.deviceMap[m] = &deviceInfo{name: name, connected: false}
		}
	}

	return nil
}

func (d *dbusWrapper) SendInitCommands(mac string, dev *deviceInfo) error {
	switch strings.ToLower(dev.name) {
	case "sensortag":
		{
			time.Sleep(1000 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA1204514000b000000000000000", "01")
			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA1304514000b000000000000000", "0A")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA1104514000b000000000000000", true)

			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA7204514000b000000000000000", "01")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA7104514000b000000000000000", true)

			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA0204514000b000000000000000", "01")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA0104514000b000000000000000", true)
		}
	case "pod":
		{
			time.Sleep(1000 * time.Millisecond)
			d.BleGattWrite(mac, "ffb2", "aa550f038401e0")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "ffb2", true)
		}
	}
	return nil
}

func (d *dbusWrapper) CloudUpdateCommand(id uint32, status string, result map[string]interface{}) {
	b, _ := json.Marshal(result)
	d.call("UpdateCommand", id, status, string(b))
}

func getAcceleration(s string) float64 {
	b, _ := hex.DecodeString(s)
	i := int8(b[0])
	j := int8(b[1])
	k := int8(b[2])

	x := float64(i) / 64.0
	y := float64(j) / 64.0
	z := float64(k) / -64.0

	// log.Printf("Acceleration (%s) [%v, %v, %v]", s, x, y, z)

	return math.Sqrt(x*x + y*y + z*z)
}

func main() {
	cloud, err := NewdbusWrapper("/com/devicehive/cloud", "com.devicehive.cloud")
	if err != nil {
		log.Panic(err)
	}

	ble, err := NewdbusWrapper("/com/devicehive/bluetooth", "com.devicehive.bluetooth")
	ble.deviceMap = make(map[string]*deviceInfo)

	if err != nil {
		log.Panic(err)
	}

	enocean, err := NewdbusWrapper("/com/devicehive/enocean", "com.devicehive.enocean")
	if err != nil {
		log.Panic(err)
	}

	myMelody := NewMelody(ble)

	ble.RegisterHandler("PeripheralDiscovered", func(args ...interface{}) {
		cloud.SendNotification("PeripheralDiscovered", map[string]interface{}{
			"mac":  args[0].(string),
			"name": args[1].(string),
			"rssi": args[2].(int16),
		}, 100)
	})

	ble.RegisterHandler("PeripheralConnected", func(args ...interface{}) {
		ble.deviceLock.Lock()
		if v, ok := ble.deviceMap[args[0].(string)]; ok {
			v.connected = true
			ble.SendInitCommands(args[0].(string), v)
		}
		ble.deviceLock.Unlock()

		cloud.SendNotification("PeripheralConnected", map[string]interface{}{
			"mac": args[0].(string),
		}, 100)
	})

	ble.RegisterHandler("PeripheralDisconnected", func(args ...interface{}) {
		ble.deviceLock.Lock()
		if v, ok := ble.deviceMap[args[0].(string)]; ok {
			v.connected = false
		}
		ble.deviceLock.Unlock()

		cloud.SendNotification("PeripheralDisconnected", map[string]interface{}{
			"mac": args[0].(string),
		}, 100)
	})

	enocean.RegisterHandler("message_received", func(args ...interface{}) {
		log.Printf("Enocean message_received: %+v", args)
		v := args[0].(string)
		var res map[string]interface{}
		err := json.Unmarshal([]byte(v), &res)

		if err != nil {
			log.Printf("Error parsing enocean response: %s", err)
			return
		}

		state := ""
		sender, ok := res["sender"].(string)

		if !ok {
			return
		}

		r1, ok := res["R1"].(map[string]interface{})

		if !ok {
			return
		}

		if int32(r1["raw_value"].(float64)) == 2 {
			state = "ON"
		}

		if int32(r1["raw_value"].(float64)) == 3 {
			state = "OFF"
		}

		cloud.SendNotification("EnoceanNotificationReceived", map[string]interface{}{
			"sender": sender,
			"state":  state,
		}, 1)
	})

	ble.RegisterHandler("NotificationReceived", func(args ...interface{}) {
		uuid := strings.ToLower(args[1].(string))
		mac := strings.ToLower(args[0].(string))
		value := strings.ToLower(args[2].(string))

		if _, ok := ble.readingsBuffer[mac]; !ok {
			ble.readingsBuffer[mac] = new([]float64)
		}

		if uuid == "f000aa1104514000b000000000000000" {
			if len(*ble.readingsBuffer[mac]) > 9 {
				vS := stats.VarS(*ble.readingsBuffer[mac])
				cloud.SendNotification("NotificationReceived", map[string]interface{}{
					"mac":   mac,
					"uuid":  uuid,
					"value": vS,
				}, 1)
				// log.Printf("Variance: [S: %v, P: %v]", vS, vP)
				ble.readingsBuffer[mac] = new([]float64)
			} else {
				r := *ble.readingsBuffer[mac]
				n := append(r, getAcceleration(value))
				ble.readingsBuffer[mac] = &n
				// log.Printf("Acceleration buffer: %v", ble.readingsBuffer[mac])
			}
		} else {
			cloud.SendNotification("NotificationReceived", map[string]interface{}{
				"mac":   mac,
				"uuid":  uuid,
				"value": value,
			}, 1)
		}

	})

	ble.RegisterHandler("IndicationReceived", func(args ...interface{}) {
		cloud.SendNotification("IndicationReceived", map[string]interface{}{
			"mac":   args[0].(string),
			"uuid":  args[1].(string),
			"value": args[2].(string),
		}, 1)
	})

	cloudHandlers := make(map[string]cloudCommandHandler)

	cloudHandlers["init"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		return nil, ble.Init(p["mac"].(string), p["type"].(string))
	}

	cloudHandlers["connect"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		t, found := p["type"] // public or random
		random := false
		if found {
			switch {
			case t.(string) == "random":
				random = true
			case t.(string) == "public":
				random = false
			default:
				return nil, errors.New("Invalid type, should be random or public")
			}
		}

		return nil, ble.BleConnect(p["mac"].(string), random)
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

	cloudHandlers["play"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		mac := p["mac"].(string)

		res, err := ble.Connected(mac)

		if err != nil {
			log.Printf("Failed to play melody with error %v", err)
			return nil, err
		}

		if !res["value"].(bool) {
			ble.BleConnect(mac, true)
		}

		return nil, myMelody.Play(mac, p["melody"].(string), int(p["bpm"].(float64)))
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
				cloud.CloudUpdateCommand(id, "success", res)
			}

		} else {
			log.Printf("Unhandled command: %s", command)
		}
	})

	// Scan for devices per
	// go func() {
	// 	for {
	// 		ble.BleScanStart()
	// 		time.Sleep(5 * time.Second)
	// 		ble.BleScanStop()
	// 		time.Sleep(10 * time.Second)
	// 	}
	// }()

	go func() {
		for {
			for mac, dev := range ble.deviceMap {
				if !dev.connected {
					err := ble.BleConnect(mac, false)
					if err != nil {
						log.Printf("Error while trying to connect: %s", err.Error())
					}
				}
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// Look for pre-configured devices
	// ble.Init("", "sensortag")
	// ble.Init("", "satechiled-0")
	// ble.Init("", "delight")
	// ble.Init("", "pod")

	// dashMac := "c84998f6a543"
	// r, _ := ble.Connected(dashMac)
	// log.Printf("BleConnected: %v", r["value"].(bool))
	// ble.BleConnect(dashMac, true)
	// r, _ = ble.Connected(dashMac)
	// log.Printf("BleConnected: %v", r["value"].(bool))
	// // time.Sleep(5 * time.Second)
	// myMelody.Play(dashMac, "8cegCegCcgCceCgec", 60)

	select {}
}
