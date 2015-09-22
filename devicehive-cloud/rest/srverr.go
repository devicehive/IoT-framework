package rest

import (
	"encoding/json"
	"fmt"
)

type Srverr struct {
	ErrorField interface{} `json:"error"`
	Message    string      `json:"notification"`
}

func (e *Srverr) Error() (t string) {
	if len(t) > 0 {
		t = t + e.Message
	}

	if e.Error != nil {
		t = fmt.Sprintf("%s (%+v)", t, e.ErrorField)
	}
	return
}

func SrverrFromJson(body []byte) error {
	var e Srverr
	if err := json.Unmarshal(body, &e); err != nil {
		return err
	} else {
		return &e
	}
}

func IsSrverr(e error) (ok bool) {
	_, ok = e.(*Srverr)
	return
}
