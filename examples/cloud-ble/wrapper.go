package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus"
)

const QueueCapacity = 2048

type dbusWrapper struct {
	conn         *dbus.Conn
	path, iface  string
	handlers     map[string]signalHandler
	handlersSync sync.Mutex

	queue *signalQueue
}

func NewdbusWrapper(path string, iface string) (*dbusWrapper, error) {
	d := new(dbusWrapper)

	conn, err := dbus.SystemBus()
	if err != nil {
		conn, err = dbus.SessionBus()
		if err != nil {
			log.Panic(err)
		}
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

func (d *dbusWrapper) CloudUpdateCommand(id uint32, status string, result map[string]interface{}) {
	b, _ := json.Marshal(result)
	d.call("UpdateCommand", id, status, string(b))
}
