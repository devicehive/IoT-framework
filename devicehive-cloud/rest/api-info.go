package rest

import "github.com/mibori/gopencils"

type ApiInfo struct {
	ApiVersion         string `json:"apiVersion"`
	ServerTimestamp    string `json:"serverTimestamp"`
	WebSocketServerUrl string `json:"webSocketServerUrl"`
}

func GetApiInfo(deviceHiveURL string) (ai ApiInfo, err error) {
	api := gopencils.Api(deviceHiveURL)
	resp := &ai
	resource, err := api.Res("info", resp).Get()

	if err == nil {
		err = resource.ProcessedError()
	}
	return
}
