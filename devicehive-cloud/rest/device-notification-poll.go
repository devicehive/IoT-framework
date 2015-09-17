package rest

import (
	"net/http"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
	"github.com/devicehive/IoT-framework/devicehive-cloud/param"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
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
) (dnrs []DeviceNotificationResource, err error) {
	api := gopencils.Api(deviceHiveURL)
	// if client != nil {
	// 	api.SetClient(client)
	// }

	resource := api.Res("device").Id(deviceGuid).Res("notification").Res("poll")
	resource.Response = &dnrs

	if requestOut != nil {
		resource.SetRequestOut(requestOut)
	}

	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}

	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	_, err = resource.Get(param.Map(params))

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}

func DeviceNotificationPollAsync(
	deviceHiveURL, deviceGuid, accessKey string,
	out chan DeviceNotificationResource, control PollAsync, // cannot be nil
) {
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	requestOut := make(chan *http.Request, 1)
	local := make(chan []DeviceNotificationResource, 1)
	isStopped := make(chan struct{})
	for {
		go func() {
			for {

				dnrs, err := DeviceNotificationPoll(deviceHiveURL, deviceGuid, accessKey, nil, client, requestOut)

				select {
				case <-isStopped:
					return
				default:
				}

				if err != nil {
					say.Infof("Polling notifications error: %s", err.Error())
					time.Sleep(time.Second)
					continue
				}

				if len(dnrs) == 0 {
					continue
				}

				local <- dnrs
				break
			}
		}()

		select {
		case dnrs := <-local:
			for _, dnr := range dnrs {
				out <- dnr
			}
			continue
		case <-control:
			select {
			case req := <-requestOut:
				isStopped <- struct{}{}
				tr.CancelRequest(req)
				return
			default:
			}
		}
	}
}
