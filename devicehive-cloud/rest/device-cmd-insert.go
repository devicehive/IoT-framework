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

	resource.Post(map[string]interface{}{
		"command": command,
	})

	// log.Printf("RESPONSE STATUS: %s", resource.Raw.Status)
	// log.Printf("RESPONSE: %+v", resource.Response)

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
