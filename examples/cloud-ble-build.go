package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/godbus/dbus"
	"log"
	"strings"
	"sync"
	"time"
)

const QueueCapacity = 2048

type dbusWrapper struct {
	conn         *dbus.Conn
	path, iface  string
	handlers     map[string]signalHandler
	handlersSync sync.Mutex

	queue     *signalQueue
	deviceMap map[string]*deviceInfo

	deviceLock sync.Mutex
}

type deviceInfo struct {
	t         string
	connected bool
}

type signalHandlerFunc func(args ...interface{})
type cloudCommandHandler func(map[string]interface{}) (map[string]interface{}, error)

type signalHandler struct {
	handler  signalHandlerFunc
	priority uint64
}

type signalItem struct {
	handler   signalHandler
	signal    *dbus.Signal
	timestamp uint64
}

type signalQueue struct {
	items     []signalItem
	queueSync sync.Mutex
	cond      *sync.Cond
}

func (q signalQueue) Len() int {
	return len(q.items)
}

func (q signalQueue) Less(i, j int) bool {
	return q.items[i].timestamp*q.items[i].handler.priority > q.items[j].timestamp*q.items[j].handler.priority
}

func (q signalQueue) Swap(i, j int) {
	if (i > len(q.items)-1) || (j > len(q.items)-1) || (i < 0) || (j < 0) {
		return
	}

	q.items[i], q.items[j] = q.items[j], q.items[i]
}

func (q *signalQueue) Push(x interface{}) {
	q.cond.L.Lock()
	q.items = append(q.items, x.(signalItem))
	q.cond.L.Unlock()
	q.cond.Signal()
}

func (q *signalQueue) Pop() interface{} {
	q.cond.L.Lock()

	old := q.items
	n := len(old)

	x := old[n-1]
	q.items = old[0 : n-1]

	q.cond.L.Unlock()
	q.cond.Signal()

	return x
}

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
	d.queue = &signalQueue{cond: &sync.Cond{L: &sync.Mutex{}}}
	heap.Init(d.queue)

	filter := fmt.Sprintf("type='signal',path='%[1]s',interface='%[2]s',sender='%[2]s'", path, iface)
	log.Printf("Filter: %s", filter)

	conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").Call("org.freedesktop.DBus.AddMatch", 0, filter)

	go func() {
		ch := make(chan *dbus.Signal, 2*QueueCapacity)
		conn.Signal(ch)
		for signal := range ch {
			if !((strings.Index(signal.Name, iface) == 0) && (string(signal.Path) == path)) {
				continue
			}
			if val, ok := d.handlers[signal.Name]; ok {
				for d.queue.Len() > QueueCapacity-1 {
					item := heap.Remove(d.queue, d.queue.Len()-1)
					log.Printf("Removing %+v from queue", item)
				}
				heap.Push(d.queue, signalItem{handler: val, signal: signal, timestamp: uint64(time.Now().Unix())})
			} else {
				log.Printf("Unhandled signal: %s", signal.Name)
			}
		}
	}()

	go func() {
		for {
			d.queue.cond.L.Lock()
			for d.queue.Len() == 0 {
				d.queue.cond.Wait()
			}
			d.queue.cond.L.Unlock()
			item := heap.Pop(d.queue).(signalItem)
			item.handler.handler(item.signal.Body...)
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
	// log.Printf("Sending cloud notification: %s", string(b))
	d.call("SendNotification", name, string(b))
}

func (d *dbusWrapper) RegisterHandler(signal string, priority uint64, h signalHandlerFunc) {
	d.handlersSync.Lock()
	d.handlers[d.iface+"."+signal] = signalHandler{priority: priority, handler: h}
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

func (d *dbusWrapper) Init(mac, t string) error {
	d.deviceLock.Lock()
	if _, ok := d.deviceMap[mac]; !ok {
		d.deviceMap[mac] = &deviceInfo{t: t, connected: false}
	}
	d.deviceLock.Unlock()
	return nil
}

func (d *dbusWrapper) SendInitCommands(mac string, dev *deviceInfo) error {
	switch strings.ToLower(dev.t) {
	case "sensortag":
		{
			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA1204514000b000000000000000", "01")
			time.Sleep(500 * time.Millisecond)
			d.BleGattWrite(mac, "F000AA1304514000b000000000000000", "0A")
			time.Sleep(500 * time.Millisecond)
			d.BleGattNotifications(mac, "F000AA1104514000b000000000000000", true)
		}
	case "pod":
		{
			time.Sleep(500 * time.Millisecond)
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

	ble.RegisterHandler("PeripheralDiscovered", 100, func(args ...interface{}) {
		cloud.SendNotification("PeripheralDiscovered", map[string]interface{}{
			"mac":  args[0].(string),
			"name": args[1].(string),
			"rssi": args[2].(int16),
		})
	})

	ble.RegisterHandler("PeripheralConnected", 100, func(args ...interface{}) {
		ble.deviceLock.Lock()
		if v, ok := ble.deviceMap[args[0].(string)]; ok {
			v.connected = true
			ble.SendInitCommands(args[0].(string), v)
		}
		ble.deviceLock.Unlock()

		cloud.SendNotification("PeripheralConnected", map[string]interface{}{
			"mac": args[0].(string),
		})
	})

	ble.RegisterHandler("PeripheralDisconnected", 100, func(args ...interface{}) {
		ble.deviceLock.Lock()
		if v, ok := ble.deviceMap[args[0].(string)]; ok {
			v.connected = false
		}
		ble.deviceLock.Unlock()

		cloud.SendNotification("PeripheralDisconnected", map[string]interface{}{
			"mac": args[0].(string),
		})
	})

	ble.RegisterHandler("NotificationReceived", 1, func(args ...interface{}) {
		cloud.SendNotification("NotificationReceived", map[string]interface{}{
			"mac":   args[0].(string),
			"uuid":  args[1].(string),
			"value": args[2].(string),
		})
	})

	ble.RegisterHandler("IndicationReceived", 1, func(args ...interface{}) {
		cloud.SendNotification("IndicationReceived", map[string]interface{}{
			"mac":   args[0].(string),
			"uuid":  args[1].(string),
			"value": args[2].(string),
		})
	})

	cloudHandlers := make(map[string]cloudCommandHandler)

	cloudHandlers["init"] = func(p map[string]interface{}) (map[string]interface{}, error) {
		return nil, ble.Init(p["mac"].(string), p["type"].(string))
	}

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

	cloud.RegisterHandler("CommandReceived", 1, func(args ...interface{}) {
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

	go func() {
		for {
			for mac, dev := range ble.deviceMap {
				if !dev.connected {
					err := ble.BleConnect(mac)
					if err != nil {
						log.Printf("Error while trying to connect: %s", err.Error())
					}
				}
				time.Sleep(1 * time.Second)
			}
		}
	}()

	select {}
}
