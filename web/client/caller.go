package client

import (
	"context"
	"net/http"
)

// Caller is just simplified endpoint invoker. Use it when you do not need to handle own http request and response.
type Caller interface {
	// Call will create decorated http request and response. Parameter (in) will be decoded into http request
	// and parameter (out) will be encoded from http response.
	Call(Option) error
}

type caller struct {
	endpoint Endpoint
}

type Option struct {
	URL      string
	Method   string
	Request  interface{}
	Response interface{}
}

func (a *caller) Call(r Option) error {
	ctx := context.WithValue(
		context.WithValue(
			context.Background(),
			"in",
			r.Request),
		"out",
		r.Response,
	)

	// prepare http request with background context. Context will help decorators
	// in wrapping extra behavior for request and response.
	req, err := http.NewRequest(r.Method, r.URL, nil)
	req = req.WithContext(ctx)
	if err != nil {
		return err
	}

	// call http resource and close body to let another calls using same endpoint tcp connection
	res, err := a.endpoint.Do(req)
	if err != nil {
		return err
	}

	return res.Body.Close()
}

// Caller decorates given endpoint with extra behavior and simplifies http client.
func NewCaller(e Endpoint, ms ...Middleware) Caller {
	for _, m := range ms {
		e = m(e)
	}

	return &caller{
		endpoint: e,
	}
}
