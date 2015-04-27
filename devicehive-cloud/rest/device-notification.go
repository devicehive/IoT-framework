package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type DeviceNotification struct {
	Id           int
	Timestamp    string
	Notification string
	Parameters   interface{}
}

//Need to Test
func deviceNotificationFromInterface(o interface{}) (dn DeviceNotification, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Could not process error %#v", r)
			}
		}
	}()

	m := o.(map[string]interface{})
	dn.Id = m["id"].(int)
	dn.Timestamp = m["timestamp"].(string)
	dn.Notification = m["notification"].(string)
	dn.Parameters = m["parameters"]
	return
}

//Need to Test
func DeviceNotificationQuery(deviceHiveURL string, deviceGuid string, parameters []Parameter) (notifications []DeviceNotification, err error) {
	requestURL, err := url.Parse(deviceHiveURL + "/device/" + deviceGuid + "/notification")
	if err != nil {
		return
	}
	IntegrateGetParameters(requestURL, parameters)

	resp, err := http.Get(requestURL.String())
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var rawNotes []interface{}
	if err = json.Unmarshal(body, &rawNotes); err != nil {
		return
	}

	for _, rawNote := range rawNotes {
		var n DeviceNotification
		n, err = deviceNotificationFromInterface(rawNote)
		if err != nil {
			return
		}
		notifications = append(notifications, n)
	}

	return
}

func DeviceNotificationGet(deviceHiveURL string, deviceGuid string, deviceId int) (n DeviceNotification, err error) {
	resp, err := http.Get(fmt.Sprintf("%s/device/%s/notification/%d", deviceHiveURL, deviceGuid, deviceId))
	if err != nil {
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var rawNote interface{}
	if err = json.Unmarshal(body, rawNote); err != nil {
		return
	}

	return deviceNotificationFromInterface(rawNote)
}
