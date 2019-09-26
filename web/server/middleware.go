package server

import (
	"encoding/json"
	"net/http"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
)

type EndpointFunc func(*Request)

func (f EndpointFunc) Do(r *Request) { f(r) }

type Endpoint interface {
	Do(*Request)
}

type Middleware func(Endpoint) Endpoint

var With = middlewares{}

type middlewares struct{}

type Logger func(message string, args ...interface{})

func (middlewares) Logger(log Logger) Middleware {
	return func(n Endpoint) Endpoint {
		return EndpointFunc(func(r *Request) {
			//n.ServeHTTP()
			u := r.Reader.URL.String()
			m := r.Reader.Method

			n.Do(r)

			log("%s %s", m, u)
		})
	}
}

func (middlewares) JSON(typ string) Middleware {
	t := transform.CamelCaseKeys(false)

	switch typ {
	case "snake":
		t = transform.ConventionalKeys()
	case "camel":
		t = transform.CamelCaseKeys(true)
	}

	return func(next Endpoint) Endpoint {
		return EndpointFunc(func(r *Request) {
			next.Do(r)

			err := r.Response.Error
			body := r.Response.Body
			res := r.Writer

			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}

			res.Header().Set("content-type", "application/json")
			res.WriteHeader(http.StatusOK)

			if err := conjson.NewEncoder(json.NewEncoder(res), t).Encode(body); err != nil {
				http.Error(r.Writer, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	}
}
