package param

import (
	"fmt"
	"net/url"
)

type I interface {
	Name() string
	String() string
}

func IntegrateWithUrl(baseURL *url.URL, params []I) {
	if params != nil && len(params) > 0 {
		q := baseURL.Query()
		for _, p := range params {
			q.Set(p.Name(), p.String())
		}
		baseURL.RawQuery = q.Encode()
	}

}

func UrlConcat(params []I) string {
	v := url.Values{}
	for _, p := range params {
		v.Add(p.Name(), p.String())
	}
	return v.Encode()
}

func Map(params []I) map[string]string {
	r := map[string]string{}
	for _, p := range params {
		r[p.Name()] = p.String()
	}
	return r
}

type Int struct {
	Arg   string
	Value int
}

func (p Int) Name() string {
	return p.Arg
}

func (p Int) String() string {
	return fmt.Sprintf("%d", p.Value)
}

type String struct {
	Arg   string
	Value string
}

func (p String) Name() string {
	return p.Arg
}

func (p String) String() string {
	return p.Value
}
