package rest

import (
	"github.com/devicehive/IoT-framework/devicehive-cloud/say"
	"testing"
	"encoding/json"
)

func TestDeviceJson(t *testing.T) {
	d := MakeDeviceInfo("my device name",
			MakeDeviceClassInfo("go-gateway-class", "0.2"))
	b, err := json.Marshal(d)
	if err != nil {
		t.Error("failed to convert device: %q", err.Error())
	}
	// t.Logf("device json:%s", string(b))
	s := `{"name":"my device name","status":"Online","deviceClass":{"name":"go-gateway-class","version":"0.2","isPermanent":false,"offlineTimeout":0}}`
	if s != string(b) {
		t.Errorf("device json format failed:\nexpected:%s\n   found:%s", s, string(b))
	}

	n := MakeNetworkInfo("my-network", "12345")
	d.SetNetwork(&n)
	b, err = json.Marshal(d)
	if err != nil {
		t.Error("failed to convert device with network: %q", err.Error())
	}
	//t.Logf("device json:%s", string(b))
	s = `{"name":"my device name","status":"Online","network":{"name":"my-network","key":"12345"},"deviceClass":{"name":"go-gateway-class","version":"0.2","isPermanent":false,"offlineTimeout":0}}`
	if s != string(b) {
		t.Errorf("device json format failed:\nexpected:%s\n   found:%s", s, string(b))
	}
}

// TestDeviceRegister() unit test for /device/{id} PUT method
func TestDeviceRegister(t *testing.T) {
	say.SetLevel(say.TRACE)

	baseURL := "http://playground.devicehive.com/api/rest" // TODO: get it from the configuration
	accessKey := "gfgCKy2KNC+MMuiwTKRTcqEPLv84gEEh9b26lQOZTq4="
	deviceId := "go-device-register-test"
	device := MakeDeviceInfo(deviceId + "-name",
			MakeDeviceClassInfo("go-gateway-class", "0.1"))
	//network := MakeNetworkInfo("Network BPGEDP for (sergey.polichnoy@dataart.com)", "D1WM8VE687QJ7LZP")
	//device.SetNetwork(&network)

	err := DeviceRegister(baseURL, deviceId, accessKey, device)
	if err != nil {
		t.Error("Failed to register device", err.Error())
	}
}
