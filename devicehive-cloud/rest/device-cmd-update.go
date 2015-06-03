package rest

import (
	"fmt"
	"net/http"

	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
)

// http://devicehive.com/restful#Reference/DeviceCommand/update

func DeviceCmdUpdate(
	deviceHiveURL, deviceGuid, accessKey string,
	commandId uint32,
	status, result string, // can be empty
) (err error) {
	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid).Res("command").Id(fmt.Sprintf("%d", commandId))

	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}
	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	requestBody := map[string]interface{}{}

	if status != "" {
		requestBody["status"] = status
	}

	if result != "" {
		requestBody["result"] = result
	}

	resource.Put(requestBody)

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
