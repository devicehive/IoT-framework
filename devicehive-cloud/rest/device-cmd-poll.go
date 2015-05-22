package rest

import (
	"net/http"

	"github.com/devicehive/IoT-framework/devicehive-cloud/param"
	"github.com/mibori/gopencils"
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

func DeviceCmdPoll(
	deviceHiveURL, deviceGuid, accessKey string,
	params []param.I, //maybe nil
	client *http.Client, //maybe nil
	requestOut chan *http.Request, //maybe nil
) (dcrs []DeviceCmdResource, err error) {
	api := gopencils.Api(deviceHiveURL)
	if client != nil {
		api.SetClient(client)
	}

	resource := api.Res("device").Id(deviceGuid).Res("command").Res("poll")
	resource.Response = &dcrs

	if requestOut != nil {
		resource.SetRequestOut(requestOut)
	}

	if resource.Headers == nil {
		resource.Headers = http.Header{}
	}
	resource.Headers["Authorization"] = []string{"Bearer " + accessKey}

	resource.Get(param.Map(params))

	// log.Printf("    DeviceCmdPoll RESPONSE STATUS: %s", resource.Raw.Status)
	// log.Printf("    DeviceCmdPoll RESPONSE: %+v", resource.Response)

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}

type PollAsync chan struct{}

func NewPollAsync() PollAsync {
	return PollAsync(make(chan struct{}, 1))
}

func (pa PollAsync) Stop() {
	pa <- struct{}{}
}

func DeviceCmdPollAsync(
	deviceHiveURL, deviceGuid, accessKey string,
	startTimestamp string, // can be empty
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

				var params []param.I
				if startTimestamp != "" {
					params = []param.I{TimestampParam(startTimestamp)}
				}

				dcrs, err := DeviceCmdPoll(deviceHiveURL, deviceGuid, accessKey, params, client, requestOut)

				select {
				case <-isStopped:
					return
				default:
				}

				if err != nil {
					continue
				}

				if len(dcrs) == 0 {
					continue
				}

				startTimestamp = func(resources []DeviceCmdResource) (maxStamp string) {
					for _, dcr := range resources {
						if dcr.Timestamp >= maxStamp {
							maxStamp = dcr.Timestamp
						}
					}
					return
				}(dcrs)

				local <- dcrs
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
