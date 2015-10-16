package rest

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

// DeviceInfo structure represents the Device object.
type DeviceInfo struct {
	Name   string `json:"name"`
	Key    string `json:"key,omitempty"`
	Status string `json:"status,omitempty"`

	Data interface{} `json:"data,omitempty"`

	Network     *NetworkInfo     `json:"network,omitempty"`
	DeviceClass *DeviceClassInfo `json:"deviceClass,omitempty"`
}


// NetworkInfo structure represents the Network object.
type NetworkInfo struct {
	Name string `json:"name"`
	Key  string `json:"key"`
	Desc string `json:"description,omitempty"`
}


// DeviceClassInfo structure represents the DeviceClass object.
type DeviceClassInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Data      interface{}     `json:"data,omitempty"`
	Equipment []EquipmentInfo `json:"equipment,omitempty"`

	IsPermanent    bool `json:"isPermanent"`
	OfflineTimeout int  `json:"offlineTimeout"`
}

// EquipmentInfo structure represents the Equipment object.
type EquipmentInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
	Type string `json:"type"`

	Data interface{} `json:"data,omitempty"`
}


// MakeDeviceInfo() function initializes the Device object.
func MakeDeviceInfo(name string, deviceClass DeviceClassInfo) (device DeviceInfo) {
	device.Name = name
	device.Status = "Online"
	device.DeviceClass = &deviceClass
	return
}

// SetNetwork() function sets the network object for the specified device
// to clear current network just pass 'nil' as an argument
func (device *DeviceInfo) SetNetwork(network *NetworkInfo) {
	device.Network = network
}

// MakeNetworkInfo() function initializes the Network object.
// Network description is empty by default
func MakeNetworkInfo(name, key string) (network NetworkInfo) {
	network.Name = name
	network.Key  = key
	return
}


// MakeDeviceClassInfo() function initializes the DeviceClass object.
func MakeDeviceClassInfo(name, version string) (deviceClass DeviceClassInfo) {
	deviceClass.Name = name
	deviceClass.Version = version
	return
}



// DeviceRegister() function registers new or existing device.
// /device/{id} PUT method is used.
// the DeviceHive server base URL should be provided as an argument
// as long as full device information.
// As a result the possible error is returned.
func DeviceRegister(baseURL, deviceId, accessKey string, device DeviceInfo) (err error) {
	say.Tracef("registering device (URL:%q, device:%+v)", baseURL, device)
	api := gopencils.Api(baseURL)

	res := api.Res("device").Id(deviceId)
	res.SetHeader("Authorization", "Bearer " + accessKey)
	res, err = res.Put(device)

	if err == nil {
		err = res.ProcessedError()
	}

	return
}
