package rest

import (
	"net/http"

	"github.com/mibori/gopencils"
)

func DeviceRegister(deviceHiveURL, deviceGuid, deviceName, accessKey string) (err error) {
	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid)

	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}
	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	params := map[string]interface{}{
		"deviceKey": "00000000-0000-0000-0000-000000000000",
		"device": map[string]interface{}{
			"key":    "00000000-0000-0000-0000-000000000000",
			"name":   deviceName,
			"status": "online",
			"network": map[string]interface{}{
				"name":        "default",
				"description": "default network"},
			"deviceClass": map[string]interface{}{
				"name":           "go-gateway-class",
				"version":        "0.1",
				"offlineTimeout": 10}}}

	resource.Put(params)

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
