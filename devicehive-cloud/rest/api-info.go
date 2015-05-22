package rest

import "github.com/mibori/gopencils"

type ApiInfo struct {
	ApiVersion         string `json:"apiVersion"`
	ServerTimestamp    string `json:"serverTimestamp"`
	WebSocketServerUrl string `json:"webSocketServerUrl"`
}

// Example from server: {ApiVersion:2.0.0 ServerTimestamp:2015-05-21T14:18:34.019584 WebSocketServerUrl:ws://52.6.240.235:8080/dh/websocket}

func GetApiInfo(deviceHiveURL string) (ai ApiInfo, err error) {
	api := gopencils.Api(deviceHiveURL)
	resp := &ai
	resource, err := api.Res("info", resp).Get()

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
