package rest

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/gopencils"
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
)

// ApiInfo structure represents the DeviceHive server's response
// See http://devicehive.com/restful/#Reference/ApiInfo/get for more details.
// Example from server: {ApiVersion:2.0.0
//   ServerTimestamp:2015-05-21T14:18:34.019584
//   WebSocketUrl:ws://52.6.240.235:8080/dh/websocket}
type ApiInfo struct {
	ApiVersion      string `json:"apiVersion"`
	ServerTimestamp string `json:"serverTimestamp"`
	WebSocketUrl    string `json:"webSocketServerUrl"`
}

// GetApiInfo() function gets the main server's information.
// /info GET method is used.
// the DeviceHive server base URL should be provided as an argument.
// As a result the the server's information is returned and possible error.
func GetApiInfo(baseURL string) (info ApiInfo, err error) {
	say.Tracef("getting server info (URL:%q)", baseURL)
	api := gopencils.Api(baseURL)
	resource, err := api.Res("info", &info).Get()
	if err == nil {
		err = resource.ProcessedError()
	}
	say.Debugf("server info (URL:%q): info=%+v, error=%q", baseURL, info, err)
	return
}
