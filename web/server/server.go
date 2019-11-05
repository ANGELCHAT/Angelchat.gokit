package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Service struct{ mux *mux.Router }

func New() *Service { return &Service{mux: mux.NewRouter()} }

func (r *Service) Prefix(prefix string, ms ...Middleware) *Service {
	pr := r.mux.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		if m == nil {
			continue
		}
		pr.Use(m)
	}

	return &Service{mux: pr}
}

func (r *Service) Handle(path string, e http.HandlerFunc, method string, ms ...Middleware) *Service {
	var h http.Handler = e
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i] == nil {
			continue
		}
		h = ms[i](h)
	}

	rh := r.mux.Handle(path, h)
	rh.Methods(method)

	return r
}

func (r *Service) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(&Request{writer: res, reader: req}, req)
}
