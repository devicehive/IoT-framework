package rest

import (
	"net/http"

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

func DeviceCmdPoll(deviceHiveURL, deviceGuid, accessKey string, client *http.Client, requestOut chan *http.Request) (dcrs []DeviceCmdResource, err error) {
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

	resource.Get()

	// log.Printf("RESPONSE STATUS: %s", resource.Raw.Status)
	// log.Printf("RESPONSE: %+v", resource.Response)

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

func DeviceCmdPollAsync(deviceHiveURL, deviceGuid, accessKey string, out chan DeviceCmdResource, control PollAsync) {
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	requestOut := make(chan *http.Request, 1)
	local := make(chan []DeviceCmdResource, 1)
	isStopped := make(chan struct{})
	for {
		go func() {
			for {
				dcrs, err := DeviceCmdPoll(deviceHiveURL, deviceGuid, accessKey, client, requestOut)

				select {
				case <-isStopped:
					return
				default:
				}

				if err != nil {
					// log.Printf("Error: %s", err.Error())
					continue
				}

				if len(dcrs) == 0 {
					continue
				}

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
			// log.Printf("-> STOP CONTROL RECEIVED")
			select {
			case req := <-requestOut:
				isStopped <- struct{}{}
				tr.CancelRequest(req)
				return
			default:
				// log.Printf("Warning: can't catch current request")
			}
		}
	}
}
