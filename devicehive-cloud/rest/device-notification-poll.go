package rest

import (
	"net/http"

	"github.com/devicehive/IoT-framework/devicehive-cloud/param"
	"github.com/mibori/gopencils"
)

type DeviceNotificationResource struct {
	Id           int         `json:"id"`
	Timestamp    string      `json:"timestamp"`
	Notification string      `json:"notification"`
	Parameters   interface{} `json:"parameters"`
}

func DeviceNotificationPoll(
	deviceHiveURL, deviceGuid, accessKey string,
	params []param.I, //maybe nil
	client *http.Client, //maybe nil
	requestOut chan *http.Request, //maybe nil
) (dnrs []DeviceCmdResource, err error) {
	api := gopencils.Api(deviceHiveURL)
	if client != nil {
		api.SetClient(client)
	}

	resource := api.Res("device").Id(deviceGuid).Res("notification").Res("poll")
	resource.Response = &dnrs

	if requestOut != nil {
		resource.SetRequestOut(requestOut)
	}

	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}
	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	resource.Get(param.Map(params))

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
