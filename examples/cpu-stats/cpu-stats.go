package main

import (
	"log"
	"os"
	"time"

	"github.com/devicehive/IoT-framework/godbus-helpers/cloud"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func main() {
	cloud, err := cloud.NewDbusForComDevicehiveCloud()
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
			}, 1)
		}
	}
}
