package main

import (
	"encoding/json"
	"github.com/godbus/dbus"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"os"
	"time"
)

type dbusWrapper struct {
	conn        *dbus.Conn
	path, iface string
}

func NewdbusWrapper(path string, iface string) (*dbusWrapper, error) {
	d := new(dbusWrapper)

	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}

	d.conn = conn
	d.path = path
	d.iface = iface

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
	d.call("SendNotification", name, string(b), uint64(1))
}

func main() {
	cloud, err := NewdbusWrapper("/com/devicehive/cloud", "com.devicehive.cloud")
	if err != nil {
		log.Panic(err)
	}

	h, _ := os.Hostname()

	for {
		time.Sleep(time.Second)
		c, err := cpu.CPUPercent(time.Second, false)
		if err != nil {
			log.Panic(err)
		}

		v, err := mem.VirtualMemory()
		if err != nil {
			log.Panic(err)
		}

		if len(c) > 0 {
			cloud.SendNotification("stats", map[string]interface{}{
				"cpu-usage":    c[0],
				"memory-total": v.Total,
				"memory-free":  v.Free,
				"name":         h,
			})
		}
	}
}
