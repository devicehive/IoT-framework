package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/godbus/dbus"
	"github.com/montanaflynn/stats"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"math"
	"os"
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
	name string
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
			if handler, ok := d.handlers[signal.Name]; ok {
				go handler(signal.Body...)
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

func (d *dbusWrapper) BleConnected(mac string) bool {
	r := false
	err := d.call("Connected", mac).Store(&r)

	if err != nil {
		log.Printf("Error while calling Connected(): %s", err)
	}

	return r
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
			d.deviceMap[m] = &deviceInfo{name: name}
		}
	}

	return nil
}

func (d *dbusWrapper) SendInitCommands(mac string, dev *deviceInfo) error {
	switch strings.ToLower(dev.name) {
	case "cc2650 sensortag":
		{
			time.Sleep(1000 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA8204514000b000000000000000", "3800")
			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA8304514000b000000000000000", "0A")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA8104514000b000000000000000", true)

			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA7204514000b000000000000000", "01")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA7104514000b000000000000000", true)

			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA0204514000b000000000000000", "01")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA0104514000b000000000000000", true)
		}
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
	case "polar-h7":
		{
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "2a37", true)
		}
	case "satechiled-0":
		{
		}
	}
	return nil
}

func (d *dbusWrapper) CloudUpdateCommand(id uint32, status string, result map[string]interface{}) {
	b, _ := json.Marshal(result)
	d.call("UpdateCommand", id, status, string(b))
}

func getAcceleration(s, uuid string) float64 {
	b, _ := hex.DecodeString(s)

	var x, y, z float64

	// f000aa11 for old sensortag
	if uuid == "f000aa1104514000b000000000000000" {
		var i, j, k int8
		i = int8(b[0])
		j = int8(b[1])
		k = int8(b[2])

		x = float64(i) / 64.0
		y = float64(j) / 64.0
		z = float64(k) / -64.0
		// f000aa81 for new sensortag
	} else if uuid == "f000aa8104514000b000000000000000" {
		var res [3]int16
		var i, j, k int16
		binary.Read(bytes.NewReader(b[6:12]), binary.LittleEndian, &res)
		i = res[0]
		j = res[1]
		k = res[2]

		x = float64(i) * 2.0 / 32768.0
		y = float64(j) * 2.0 / 32768.0
		z = float64(k) * 2.0 / 32768.0
	} else {
		log.Printf("Error: unknown uuid to read accelerometer data from: %s", uuid)
	}

	// log.Printf("Acceleration (%s) [%v, %v, %v]", s, x, y, z)

	return math.Sqrt(x*x + y*y + z*z)
}

func parseHRate(s string) int {
	b, _ := hex.DecodeString(s)

	val := int(b[1])
	if b[0] & 0x01 == 0x01 {
		val += int(b[2] << 8)
	}
	return val
}

type Conf struct {
	LedMac             string `yaml:"LedMac,omitempty"`
	HeartRateSensorMac string `yaml:"HeartRateSensorMac,omitempty"`
	HighHeartRate      int    `yaml:"HighHeartRate,omitempty"`
}

func main() {
	// parse command-line args
	confFile := ""
	flag.StringVar(&confFile, "conf", "", "YAML configuration for this demo")
	if !flag.Parsed() {
		flag.Parse()
	}

	sampleConf := Conf{}
	sampleConf.HeartRateSensorMac = "112233445566"
	sampleConf.LedMac = "665544332211"
	sampleConf.HighHeartRate = 75
	sampleYaml, _ := yaml.Marshal(sampleConf)

	if confFile == "" {
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Sample config:\n")
		fmt.Printf("%s", sampleYaml)
		return
	}

	conf := Conf{}

	// parse config
	yamlFile, err := ioutil.ReadFile(confFile)
	if err != nil {
		log.Panic(err)
	}
	err = yaml.Unmarshal(yamlFile, &conf)
	if err != nil {
		log.Panic(err)
	}

	// set defaults
	if conf.HighHeartRate == 0 {
		// value above which we switch to red
		conf.HighHeartRate = 80
	}
	if conf.LedMac == "" || conf.HeartRateSensorMac == ""  {
		fmt.Fprintf(os.Stderr, "LedMac or HeartRateSensorMac not set in config; sample config:\n")
		fmt.Printf("%s", sampleYaml)
		return
	}

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

	// pulse LED based on heart rate; default of 0 disables until
	// we find a heart rate sensor
	heartRate := 0
	ledMac := conf.LedMac
	heartSensorMac := conf.HeartRateSensorMac
	go func() {
		for {
			if heartRate > 0 && ble.BleConnected(ledMac) {
				log.Printf("heartRate=%d", heartRate)
				pulsePeriod := time.Duration(60 * 1000 / heartRate) * time.Millisecond
				//log.Printf("pulsePeriod=%s, full=%s, dim=%s", pulsePeriod, pulsePeriod * 2/3, pulsePeriod * 1/3)
				RED := "0f0d0300ff00006400000000000067ffff"
				DIMRED := "0f0d0300ff00001e00000000000021ffff"
				GREEN := "0f0d030000ff006400000000000067ffff"
				DIMGREEN := "0f0d030000ff001e00000000000021ffff"
				if (heartRate > conf.HighHeartRate) {
					// RED
					ble.BleGattWrite(ledMac, "fff3", RED)
					time.Sleep(pulsePeriod * 2/3)
					// DIM RED
					ble.BleGattWrite(ledMac, "fff3", DIMRED)
					time.Sleep(pulsePeriod * 1/3)
				} else {
					// GREEN
					ble.BleGattWrite(ledMac, "fff3", GREEN)
					time.Sleep(pulsePeriod * 2/3)
					// DIM GREEN
					ble.BleGattWrite(ledMac, "fff3", DIMGREEN)
					time.Sleep(pulsePeriod * 1/3)
				}
			} else {
				time.Sleep(3 * time.Second)
			}
		}
	}()

	ble.RegisterHandler("PeripheralDiscovered", func(args ...interface{}) {
		mac := strings.ToLower(args[0].(string))
		name := strings.ToLower(args[1].(string))

		log.Printf("Discovered new device mac: %s, name: %s", mac, name)
		if _, ok := ble.scanMap[name]; ok {
			if _, ok := ble.deviceMap[mac]; !ok {
				log.Printf("Discovered new device [mac: %s, name: %s]", mac, name)
				ble.deviceLock.Lock()
				ble.deviceMap[mac] = &deviceInfo{name: name}
				ble.deviceLock.Unlock()
			}
		}

		log.Printf("PeripheralDiscovered mac=%s, name=%s", args[0].(string), args[1].(string))
		cloud.SendNotification("PeripheralDiscovered", map[string]interface{}{
			"mac":  args[0].(string),
			"name": args[1].(string),
			"rssi": args[2].(int16),
		}, 100)
	})

	ble.RegisterHandler("PeripheralConnected", func(args ...interface{}) {
		ble.deviceLock.Lock()
		if v, ok := ble.deviceMap[args[0].(string)]; ok {
			log.Printf("sending init commands for mac=%s, name=%s", args[0].(string), v.name)
			ble.SendInitCommands(args[0].(string), v)
		}
		ble.deviceLock.Unlock()

		log.Printf("PeripheralConnected mac=%s", args[0].(string))
		cloud.SendNotification("PeripheralConnected", map[string]interface{}{
			"mac": args[0].(string),
		}, 100)
	})

	ble.RegisterHandler("PeripheralDisconnected", func(args ...interface{}) {
		log.Printf("PeripheralDisconnected mac=%s", args[0].(string))
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

		log.Printf("NotificationReceived mac=%s, uuid=%s, value=%s", mac, uuid, value)

		if _, ok := ble.readingsBuffer[mac]; !ok {
			ble.readingsBuffer[mac] = new([]float64)
		}

		// f000aa11 for old sensortag, f000aa81 for new sensortag
		if uuid == "f000aa1104514000b000000000000000" || uuid == "f000aa8104514000b000000000000000" {
			if len(*ble.readingsBuffer[mac]) > 9 {
				vS, _ := stats.VarS(*ble.readingsBuffer[mac])
				cloud.SendNotification("NotificationReceived", map[string]interface{}{
					"mac":   mac,
					"uuid":  uuid,
					"value": vS,
				}, 1)
				// log.Printf("Variance: [S: %v, P: %v]", vS, vP)
				ble.readingsBuffer[mac] = new([]float64)
			} else {
				r := *ble.readingsBuffer[mac]
				n := append(r, getAcceleration(value, uuid))
				ble.readingsBuffer[mac] = &n
				// log.Printf("Acceleration buffer: %v", ble.readingsBuffer[mac])
			}
		} else {
			// heart rate measurement
			if uuid == "2a37" {
				heartRate = parseHRate(value)
			}

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

		if !ble.BleConnected(mac) {
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

	// do a single scan on startup
	log.Printf("Starting scan...")
	ble.BleScanStart()
	time.Sleep(5 * time.Second)
	log.Printf("Settling after scan...")
	ble.BleScanStop()
	time.Sleep(1 * time.Second)

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
			for mac, _ := range ble.deviceMap {
				if !ble.BleConnected(mac) {
					log.Printf("Connecting mac=%s", mac)
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
	// ble.Init("68c90b047306", "cc2650 sensortag")
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

	// prefix logs with short filename
	log.SetFlags(log.Flags() | log.Lshortfile)

	// connect just to this particular LED; NB: device name is actually
	// SATECHILED-0
	ble.Init(ledMac, "satechiled-0")
	// this should connect to all LEDs of this type, but no
	// PeripheralConnected is triggered in this case
	//ble.Init("", "SATECHILED-0")

	// reported device name is empty; looking manually at GAP's 0x2a00, it
	// is "Polar H7 637B4E12"; just use polar-h7
	ble.Init(heartSensorMac, "polar-h7")

	select {}
}
