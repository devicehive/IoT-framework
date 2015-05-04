package rest

import (
	"bytes"
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

	if id, ok := m["id"]; ok {
		dn.Id = id.(int)
	}

	if timestamp, ok := m["timestamp"]; ok {
		dn.Timestamp = timestamp.(string)
	}

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

// Need to test
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

// Need to test
func DeviceNotificationInsert(deviceHiveURL string, deviceGuid, notification string, parameters interface{}) (n DeviceNotification, err error) {

	n.Notification = notification
	n.Parameters = parameters

	url := fmt.Sprintf("%s/device/%s/notification", deviceHiveURL, deviceGuid)
	jsonData := map[string]interface{}{"notification": notification}
	if parameters != nil {
		jsonData["parameters"] = parameters
	}
	jsonStr, err := json.Marshal(jsonData)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	//req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var rawSemiNote interface{}
	if err = json.Unmarshal(body, rawSemiNote); err != nil {
		return
	}

	return deviceNotificationFromInterface(rawSemiNote)

}

type PollingCtx chan struct{}

func newPollingCtx() PollingCtx {
	return PollingCtx(make(chan struct{}, 1))
}

func (p PollingCtx) Cancel() {
	p <- struct{}{}
}

// Need to test
func DeviceNotificationPollAsync(deviceHiveURL string, deviceGuid string, parameters []Parameter,
	completion func(n DeviceNotification, err error, interrupted bool)) (ctx PollingCtx, err error) {

	type answer struct {
		n DeviceNotification
		e error
	}

	ctx = newPollingCtx()

	requestURL, err := url.Parse(fmt.Sprintf("%s/device/%s/notification/poll", deviceHiveURL, deviceGuid))
	if err != nil {
		return
	}
	IntegrateGetParameters(requestURL, parameters)

	req, err := http.NewRequest("GET", requestURL.String(), nil)
	if err != nil {
		return
	}
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}

	serverChan := make(chan answer, 1)

	go func() {
		resp, e := client.Do(req)
		if e != nil {
			serverChan <- answer{e: e}
			return
		}
		defer resp.Body.Close()

		body, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			return
		}

		var rawNote interface{}
		if e = json.Unmarshal(body, rawNote); e != nil {
			return
		}

		n, e := deviceNotificationFromInterface(rawNote)
		serverChan <- answer{n, e}
	}()

	go func() {
		select {
		case <-ctx:
			tr.CancelRequest(req)
			completion(DeviceNotification{}, nil, true)
		case ans := <-serverChan:
			completion(ans.n, ans.e, false)
		}
	}()
	return
}
