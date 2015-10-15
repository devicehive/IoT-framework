package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/devicehive/IoT-framework/devicehive-cloud/param"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

type DeviceNotificationResource struct {
	Id           int         `json:"id"`
	Timestamp    string      `json:"timestamp"`
	Notification string      `json:"notification"`
	Parameters   interface{} `json:"parameters"`
}

/*

Ubuntu 14.0.4 LTE. File leak version:

2015/09/22 11:55:14 INFO:Polling notifications error: Get http://nn8615.pg.devicehive.com/api/device/0B24431A-EC99-4887-8B4F-38C3CEAF1D03/notification/poll: dial tcp: lookup nn8615.pg.devicehive.com: too many open files
2015/09/22 11:55:15 INFO:Polling notifications error: Get http://nn8615.pg.devicehive.com/api/device/0B24431A-EC99-4887-8B4F-38C3CEAF1D03/notification/poll: dial tcp: lookup nn8615.pg.devicehive.com: too many open files
2015/09/22 11:55:16 INFO:Polling notifications error: Get http://nn8615.pg.devicehive.com/api/device/0B24431A-EC99-4887-8B4F-38C3CEAF1D03/notification/poll: dial tcp: lookup nn8615.pg.devicehive.com: too many open files



func DeviceNotificationPoll(
	deviceHiveURL, deviceGuid, accessKey string,
	params []param.I, //maybe nil
	client *http.Client, //maybe nil
	requestOut chan *http.Request, //maybe nil
) (dnrs []DeviceNotificationResource, err error) {
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

	_, err = resource.Get(param.Map(params))

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}

*/

func DeviceNotificationPoll(
	deviceHiveURL, deviceGuid, accessKey string,
	params []param.I, //maybe nil
	client *http.Client, //maybe nil
	requestOut chan *http.Request, //maybe nil
) (dnrs []DeviceNotificationResource, err error) {
	url := fmt.Sprintf("%s/device/%s/notification/poll", deviceHiveURL, deviceGuid)
	if client == nil {
		client = http.DefaultClient
	}

	url = param.IntegrateWithUrl(url, params)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	request.Header.Set("Authorization", "Bearer "+accessKey)

	if requestOut != nil {
		select {
		case requestOut <- request:
		default:
			say.Debugf("You use requestout chan, but this chan is full.")
		}
	}

	say.Debugf("Starting request %+v", say.RequestStr(request))
	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	if body != nil && len(body) > 0 {
		say.Debugf("Request %s received response body: %s", say.RequestStr(request), string(body))
	} else {
		say.Debugf("Request %s received zero body", say.RequestStr(request))
	}

	err = json.Unmarshal(body, &dnrs)
	if err != nil {
		if e := SrverrFromJson(body); IsSrverr(e) {
			err = e
		}
		return
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
				case <-requestOut:
				default:
				}

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
