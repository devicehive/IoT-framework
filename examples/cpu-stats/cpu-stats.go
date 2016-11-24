/*
  DeviceHive IoT-Framework business logic

  Copyright (C) 2016 DataArt

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at
 
      http://www.apache.org/licenses/LICENSE-2.0
 
  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/ 

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
