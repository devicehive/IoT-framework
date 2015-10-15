package rest

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
	"testing"
)


// TestGetApiInfo() unit test for /info method
func TestGetApiInfo(t *testing.T) {
	say.SetLevel(say.TRACE)

	baseURL := "http://playground.devicehive.com/api/rest" // TODO: get it from the configuration
	info, err := GetApiInfo(baseURL)
	if err != nil {
		t.Error("Failed to get server info", err.Error())
	}

	t.Log("/info:", info)
	if len(info.ApiVersion) == 0 {
		t.Error("No API version")
	}

	if len(info.ServerTimestamp) == 0 {
		t.Error("No server timestamp")
	}

	// websocket URL might be empty
}


// TestGetApiInfoBadAddress() unit test for /info method (invalid server address)
func TestGetApiInfoBadAddress(t *testing.T) {
	say.SetLevel(say.TRACE)

	baseURL := "http://playZZZround.devicehive.com/api/rest" // TODO: get it from the configuration
	_, err := GetApiInfo(baseURL)
	if err == nil {
		t.Error("Expected 'unknown host' error")
	}
}


// TestGetApiInfoBadPath() unit test for /info method (invalid path)
func TestGetApiInfoBadPath(t *testing.T) {
	say.SetLevel(say.TRACE)

	baseURL := "http://playground.devicehive.com/api/reZt" // TODO: get it from the configuration
	_, err := GetApiInfo(baseURL)
	if err == nil {
		t.Error("Expected 'invalid path' error")
	}
}
