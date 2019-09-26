package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
)

var With = middlewares{}

type Logger func(message string, args ...interface{})

type EndpointFunc func(*Request)

func (f EndpointFunc) Do(r *Request) { f(r) }

type Endpoint interface {
	Do(*Request)
}

type Middleware func(Endpoint) Endpoint

type middlewares struct{}

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

func (middlewares) Test(label string) Middleware {
	return func(n Endpoint) Endpoint {
		return EndpointFunc(func(r *Request) {
			fmt.Printf("%s: #1\n", label)
			{
				n.Do(r)
			}
			fmt.Printf("%s: #2\n", label)
		})
	}
}

// Mtoh convert server.Middleware to standard http Middleware
func Mtoh(m Middleware) func(http.Handler) http.Handler {
	return func(n http.Handler) http.Handler {
		f := EndpointFunc(func(r *Request) { n.ServeHTTP(r.Writer, r.Reader) })
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			m(f).Do(requestGet(res, req))
		})
	}
}

// Htom converts standard http Middleware to server.Middleware
func Htom(m func(http.Handler) http.Handler) Middleware {
	return func(e Endpoint) Endpoint {
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e.Do(requestGet(w, r))
		})
		return EndpointFunc(func(r *Request) {
			m(f).ServeHTTP(r.Writer, r.Reader)
		})
	}
}
