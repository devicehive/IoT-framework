package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type DHServerInfo struct {
	ApiVersion         string
	ServerTimestamp    string
	WebSocketServerUrl string
}

func GetDHServerInfo(deviceHiveURL string) (dh DHServerInfo, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Could not process error %#v", r)
			}
		}
	}()

	resp, err := http.Get(deviceHiveURL + "/info")
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var dat map[string]interface{}
	if err = json.Unmarshal(body, &dat); err != nil {
		return
	}

	return DHServerInfo{
		ApiVersion:         dat["apiVersion"].(string),
		ServerTimestamp:    dat["serverTimestamp"].(string),
		WebSocketServerUrl: dat["webSocketServerUrl"].(string),
	}, nil
}
