package rest

import (
	"net/http"

	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
)

func DeviceRegisterEasy(deviceHiveURL, deviceGuid, deviceName string) (err error) {
	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid)

	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}

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
