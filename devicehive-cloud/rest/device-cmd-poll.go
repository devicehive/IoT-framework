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

type DeviceCmdResource struct {
	Id         int         `json:"id"`
	Timestamp  string      `json:"timestamp"`
	UserId     int         `json:"userId"`
	Command    string      `json:"command"`
	Parameters interface{} `json:"parameters"`
	Lifetime   int         `json:"lifetime"`
	Status     string      `json:"status"`
	Result     interface{} `json:"result"`
}

// func DeviceCmdPoll(
// 	deviceHiveURL, deviceGuid, accessKey string,
// 	params []param.I, //maybe nil
// 	client *http.Client, //maybe nil
// 	requestOut chan *http.Request, //maybe nil
// ) (dcrs []DeviceCmdResource, err error) {
// 	api := gopencils.Api(deviceHiveURL)
// 	if client != nil {
// 		api.SetClient(client)
// 	}

// 	resource := api.Res("device").Id(deviceGuid).Res("command").Res("poll")
// 	resource.Response = &dcrs

// 	if requestOut != nil {
// 		resource.SetRequestOut(requestOut)
// 	}

// 	if resource.Headers == nil {
// 		resource.Headers = http.Header{}
// 	}
// 	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

// 	resource.Get(param.Map(params))

// 	// log.Printf("    DeviceCmdPoll RESPONSE STATUS: %s", resource.Raw.Status)
// 	// log.Printf("    DeviceCmdPoll RESPONSE: %+v", resource.Response)

// 	if err == nil {
// 		err = resource.ProcessedError()
// 	}
// 	return
// }

func DeviceCmdPoll(
	deviceHiveURL, deviceGuid, accessKey string,
	params []param.I, //maybe nil
	client *http.Client, //maybe nil
	requestOut chan *http.Request, //maybe nil
) (dcrs []DeviceCmdResource, err error) {
	url := fmt.Sprintf("%s/device/%s/command/poll", deviceHiveURL, deviceGuid)
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

	err = json.Unmarshal(body, &dcrs)
	if err != nil {
		if e := SrverrFromJson(body); IsSrverr(e) {
			err = e
		}
		return
	}

	return
}

// func DeviceCmdPollAsync(
// 	deviceHiveURL, deviceGuid, accessKey string,
// 	// startTimestamp string, // can be empty
// 	out chan DeviceCmdResource, control PollAsync, // cannot be nil
// ) {
// 	tr := &http.Transport{}
// 	client := &http.Client{Transport: tr}

// 	requestOut := make(chan *http.Request, 1)
// 	local := make(chan []DeviceCmdResource, 1)
// 	isStopped := make(chan struct{})
// 	for {
// 		go func() {
// 			for {

// 				// var params []param.I
// 				// if startTimestamp != "" {
// 				// 	params = []param.I{TimestampParam(startTimestamp)}
// 				// }

// 				// dcrs, err := DeviceCmdPoll(deviceHiveURL, deviceGuid, accessKey, params, client, requestOut)
// 				dcrs, err := DeviceCmdPoll(deviceHiveURL, deviceGuid, accessKey, nil, client, requestOut)

// 				select {
// 				case <-isStopped:
// 					return
// 				default:
// 				}

// 				if err != nil {
// 					continue
// 				}

// 				if len(dcrs) == 0 {
// 					continue
// 				}

// 				// startTimestamp = func(resources []DeviceCmdResource) (maxStamp string) {
// 				// 	for _, dcr := range resources {
// 				// 		if dcr.Timestamp >= maxStamp {
// 				// 			maxStamp = dcr.Timestamp
// 				// 		}
// 				// 	}
// 				// 	return
// 				// }(dcrs)

// 				local <- dcrs
// 				break
// 			}
// 		}()

// 		select {
// 		case dcrs := <-local:
// 			for _, dcr := range dcrs {
// 				out <- dcr
// 			}
// 			continue
// 		case <-control:
// 			select {
// 			case req := <-requestOut:
// 				isStopped <- struct{}{}
// 				tr.CancelRequest(req)
// 				return
// 			default:
// 			}
// 		}
// 	}
// }

func DeviceCmdPollAsync(
	deviceHiveURL, deviceGuid, accessKey string,
	out chan DeviceCmdResource, control PollAsync, // cannot be nil
) {
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	requestOut := make(chan *http.Request, 1)
	local := make(chan []DeviceCmdResource, 1)
	isStopped := make(chan struct{})
	for {
		go func() {
			for {

				dnrs, err := DeviceCmdPoll(deviceHiveURL, deviceGuid, accessKey, nil, client, requestOut)
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
					say.Infof("Polling commands error: %s", err.Error())
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
		case dcrs := <-local:
			for _, dcr := range dcrs {
				out <- dcr
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
