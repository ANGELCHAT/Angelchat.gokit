package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
)

var With middleware

type Logger func(message string, args ...interface{})

type middleware struct{}

func (middleware) Logger(log Logger) Middleware {
	return func(n Endpoint) Endpoint {
		return EndpointFunc(func(r *Request) {
			//n.ServeHTTP()
			u := r.Reader.URL.String()
			m := r.Reader.Method

			n.Do(r)

			log("%s [%d] %s", m, r.Response.Status, u)
		})
	}
}

func (middleware) JSON(typ string) Middleware {
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

			if r.Response.Error != nil {
				return
			}

			w := r.Writer
			w.Header().Set("content-type", "application/json")

			if err := conjson.NewEncoder(json.NewEncoder(w), t).Encode(r.Response.Body); err != nil {
				r.Response.Error = err
				r.Response.Status = http.StatusInternalServerError
			}

		})
	}
}

func (middleware) Error(f func(error) (string, int)) Middleware {
	if f == nil {
		f = func(err error) (string, int) { return err.Error(), http.StatusBadRequest }
	}

	return func(next Endpoint) Endpoint {
		return EndpointFunc(func(r *Request) {
			next.Do(r)
			if r.Response.Error != nil {
				var message string
				message, r.Response.Status = f(r.Response.Error)
				http.Error(r.Writer, message, r.Response.Status)
				return
			}
			r.Response.Status = http.StatusOK
		})
	}
}

func (middleware) Test(label string) Middleware {
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
