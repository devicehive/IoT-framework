package rest

import (
	"fmt"
	"net/url"
)

// Abstraction for optional named parameters
type Parameter interface {
	Name() string
}

// Integrate parameters to the url
func IntegrateGetParameters(baseURL *url.URL, parameters []Parameter) {
	q := baseURL.Query()
	for _, parameter := range parameters {
		switch parameter.(type) {
		case StringParameter:
			q.Set(parameter.Name(), parameter.(StringParameter).Value)
		case IntParameter:
			q.Set(parameter.Name(), fmt.Sprintf("%d", parameter.(IntParameter).Value))
		}
	}
	baseURL.RawQuery = q.Encode()
}

// Int parameter
type IntParameter struct {
	Arg   string
	Value int
}

func (p IntParameter) Name() string {
	return p.Arg
}

// String parameter
type StringParameter struct {
	Arg   string
	Value string
}

func (p StringParameter) Name() string {
	return p.Arg
}
