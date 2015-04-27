package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DHServerInfo struct {
	ApiVersion         string
	ServerTimestamp    string
	WebSocketServerUrl string
}

func GetDHServerInfo(deviceHiveURL string) (DHServerInfo, error) {
	resp, err := http.Get(deviceHiveURL + "/info")
	if err != nil {
		return DHServerInfo{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return DHServerInfo{}, err
	}

	var dat map[string]interface{}
	if err = json.Unmarshal(body, &dat); err != nil {
		return DHServerInfo{}, err
	}

	return DHServerInfo{
		ApiVersion:         dat["apiVersion"].(string),
		ServerTimestamp:    dat["serverTimestamp"].(string),
		WebSocketServerUrl: dat["webSocketServerUrl"].(string),
	}, nil
}
