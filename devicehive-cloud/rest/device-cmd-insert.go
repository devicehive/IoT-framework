package rest

import (
	"net/http"

	"github.com/mibori/gopencils"
)

type DeviceCmdInsertResponse struct {
	Id        int    `json:"id"`
	Timestamp string `json:"timestamp"`
	UserId    int    `json:"userId"`
}

func DeviceCmdInsert(deviceHiveURL, deviceGuid, accessKey, command string, parameters interface{}) (dcir DeviceCmdInsertResponse, err error) {
	api := gopencils.Api(deviceHiveURL)

	resource := api.Res("device").Id(deviceGuid).Res("command")
	resource.Response = &dcir
	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}
	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	requestBody := map[string]interface{}{
		"command": command,
	}

	if parameters != nil {
		requestBody["parameters"] = parameters
	}

	resource.Post(requestBody)

	// log.Printf("DeviceCmdInsert RESPONSE STATUS: %s", resource.Raw.Status)
	// log.Printf("DeviceCmdInsert RESPONSE: %+v", resource.Response)

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
