package rest

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"

)

func DeviceRegisterEasy(deviceHiveURL, deviceGuid, accessKey, deviceName, deviceKey, networkName, networkKey, networkDesc string) (err error) {
	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid)
	resource.SetHeader("Authorization", "Bearer "+accessKey)

	// device class
	dc := map[string]interface{}{
		"name":           "go-gateway-class",
		"version":        "0.1",
		"offlineTimeout": 10}

	// network (optional)
	n := map[string]interface{}{
		"name":        networkName,
		"key":         networkKey,
		"description": networkDesc}

	// device
	d := map[string]interface{}{
		// [optional] "key":    deviceKey,
		"name":   deviceName,
		"status": "Online",
		"deviceClass": dc}

	// omit "network" if all fields are empty
	if len(networkName)!=0 || len(networkKey)!=0 || len(networkDesc)!=0 {
		d["network"] = n
	}

	// omit device key if empty
	if len(deviceKey)!=0 {
		d["key"] = deviceKey
	}

	resp, err2 := resource.Put(d)
	say.Infof("DeviceRegisterEasy Response: %+v \r\n %+v", resp, resp.Raw)
		


	if err2 == nil {
		err = resource.ProcessedError()
	}
	return
}
