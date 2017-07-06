package http

import "net/http"

var Default Endpoint = http.DefaultClient

// Endpoint is http client, which can send and receive http Request and http.Response
type Endpoint interface {
	Do(r *http.Request) (*http.Response, error)
}

// EndpointFunc
type EndpointFunc func(*http.Request) (*http.Response, error)

func (f EndpointFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}
