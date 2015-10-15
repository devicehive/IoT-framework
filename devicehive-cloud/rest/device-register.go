package rest

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
)

func DeviceRegisterEasy(deviceHiveURL, deviceGuid, accessKey, deviceName string) (err error) {
	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid)
	resource.SetHeader("Authorization", "Bearer "+accessKey)

	params := map[string]interface{}{
		// "key":    "00000000-0000-0000-0000-000000000000",
		"name":   deviceName,
		"status": "online",
		// "network": map[string]interface{}{
		// 	"name":        "default",
		// 	"description": "default network",
		// },
		"deviceClass": map[string]interface{}{
			"name":           "go-gateway-class",
			"version":        "0.1",
			"offlineTimeout": 10,
		},
	}

	_, err = resource.Put(params)

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
