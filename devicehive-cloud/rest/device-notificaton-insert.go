package rest

import (
	"net/http"

	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
)

type DeviceNotificationInsertResponse struct {
	Id        int    `json:"id"`
	Timestamp string `json:"timestamp"`
}

func DeviceNotificationInsert(deviceHiveURL, deviceGuid, accessKey, name string, parameters interface{}) (dnir DeviceNotificationInsertResponse, err error) {

	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid).Res("notification")
	resource.Response = &dnir
	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}
	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	requestBody := map[string]interface{}{"notification": name}
	if parameters != nil {
		requestBody["parameters"] = parameters
	}

	resource.Post(requestBody)
	if err == nil {
		err = resource.ProcessedError()
	}

	return

}
