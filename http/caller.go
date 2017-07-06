package http

import (
	"context"
	"net/http"
)

// Caller is just simplified endpoint invoker. Use it when you do not need to handle own http request and response.
type Caller interface {
	// Call will create decorated http request and response. Parameter (in) will be decoded into http request
	// and parameter (out) will be encoded from http response.
	Call(in interface{}, out interface{}) error
}

type caller struct {
	endpoint Endpoint
}

func (a *caller) Call(in interface{}, out interface{}) error {

	ctx := context.WithValue(
		context.WithValue(
			context.Background(),
			"in",
			in),
		"out",
		out,
	)

	// prepare http request with background context. Context will help decorators
	// in wrapping extra behavior for request and response.
	req, err := http.NewRequest("", "", nil)
	req = req.WithContext(ctx)
	if err != nil {
		return err
	}

	// call http resource and close body to let another calls using same endpoint tcp connection
	res, err := a.endpoint.Do(req)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

// Caller decorates given endpoint with extra behavior and simplifies http client.
func NewCaller(e Endpoint, mc ...Wrapper) Caller {
	for _, m := range mc {
		e = m(e)
	}

	return &caller{
		endpoint: e,
	}
}
