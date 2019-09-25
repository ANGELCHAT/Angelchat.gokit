package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct{ r *mux.Router }

func NewRouter() *Router {
	return &Router{
		r: mux.NewRouter(),
	}
}

func (r *Router) Prefix(prefix string, ms ...Middleware) *Router {
	pr := r.r.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		pr.Use(mux.MiddlewareFunc(m))
	}

	return &Router{r: pr}
}

func (r *Router) Handle(path string, h http.HandlerFunc, method string, ms ...Middleware) *Router {
	handler := http.Handler(h)
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}

	rh := r.r.Handle(path, handler)
	rh.Methods(method)

	return r
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.r.ServeHTTP(res, req)
}
