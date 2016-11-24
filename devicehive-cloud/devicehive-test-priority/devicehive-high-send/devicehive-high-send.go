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

	"github.com/devicehive/IoT-framework/godbus-helpers/cloud"
)

func main() {
	c, err := cloud.NewDbusForComDevicehiveCloud()
	if err != nil {
		log.Fatalf("Creation Dbus wrapper with error: %s", err.Error())
	}

	c.SendNotification("[H I G H  N O T I F I C A T I O N]", map[string]interface{}{}, 1000)
}
