package server

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
)

var With middleware

type (
	Logger     func(message string, args ...interface{})
	Middleware = func(http.Handler) http.Handler
)

type middleware struct{}

func (middleware) Logger(log Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//n.ServeHTTP()
			u := r.URL.String()
			m := r.Method

			next.ServeHTTP(w, r)

			if w, ok := w.(*Request); ok {
				log("%s [%d] %s", m, w.status, u)
			}

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

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			rw, ok := w.(*Request)
			if !ok {
				return
			}

			if rw.err != nil || rw.status >= 400 {
				return
			}

			var b bytes.Buffer
			rw.err = conjson.NewEncoder(json.NewEncoder(&b), t).Encode(rw.body)
			if rw.err != nil {
				return
			}

			rw.Header().Set("content-type", "application/json")
			rw.WriteHeader(http.StatusOK)
			_, rw.err = b.WriteTo(rw)
		})
	}
}

func (middleware) Error(f func(error) (string, int)) Middleware {
	if f == nil {
		f = func(err error) (string, int) { return err.Error(), http.StatusBadRequest }
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			rw, ok := w.(*Request)
			if !ok {
				return
			}

			if rw.err != nil && rw.status == 0 {
				message, status := f(rw.err)
				http.Error(rw, message, status)
				return
			}
		})
	}
}
