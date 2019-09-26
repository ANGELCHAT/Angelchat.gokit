package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct{ mux *mux.Router }

func New() *Router { return &Router{mux: mux.NewRouter()} }

func (r *Router) Prefix(prefix string, ms ...Middleware) *Router {
	pr := r.mux.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		pr.Use(Mtoh(m))
	}

	return &Router{mux: pr}
}

func (r *Router) Handle(path string, e EndpointFunc, method string, ms ...Middleware) *Router {
	var h Endpoint = e
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}

	f := func(res http.ResponseWriter, req *http.Request) { h.Do(requestGet(res, req)) }
	rh := r.mux.Handle(path, http.HandlerFunc(f))
	rh.Methods(method)

	return r
}

func (r *Router) Do(req *Request) {}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(res, requestSet(res, req))
}

func requestGet(res http.ResponseWriter, req *http.Request) *Request {
	x := req.Context().Value(&rkey).(*Request)
	x.Reader = req
	x.Writer = res
	return x
}

func requestSet(res http.ResponseWriter, req *http.Request) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), &rkey, &Request{}))
}

var rkey = "covered-request"
