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

package cloud

import (
	"encoding/json"

	"github.com/godbus/dbus"

	"github.com/devicehive/IoT-framework/godbus-helpers/dbushelper"
)

type Dbus struct{ *dbushelper.Dbus }

func NewDbus(iface, path string) (*Dbus, error) {
	base, err := dbushelper.NewDbus(iface, path)
	return &Dbus{base}, err
}

func NewDbusForComDevicehiveCloud() (*Dbus, error) {
	return NewDbus(PathComDevicehiveCloud, IfaceComDevicehiveCloud)
}

func (w *Dbus) SendNotification(name, parameters interface{}, priority uint64) (*dbus.Call, error) {
	b, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	c := w.Call("SendNotification", name, string(b), priority)
	return c, c.Err
}

const (
	PathComDevicehiveCloud  = "/com/devicehive/cloud"
	IfaceComDevicehiveCloud = "com.devicehive.cloud"
)
