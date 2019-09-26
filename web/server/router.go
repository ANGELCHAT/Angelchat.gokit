package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Router struct{ mux *mux.Router }

func NewRouter() *Router { return &Router{mux: mux.NewRouter()} }

func (r *Router) Prefix(prefix string, ms ...Middleware) *Router {
	pr := r.mux.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		pr.Use(r.mtom(m))
	}

	return &Router{mux: pr}
}

func (r *Router) Handle(path string, e EndpointFunc, method string, ms ...Middleware) *Router {
	var h Endpoint = e
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}

	f := func(res http.ResponseWriter, req *http.Request) { h.Do(r.request(res, req)) }
	rh := r.mux.Handle(path, http.HandlerFunc(f))
	rh.Methods(method)

	return r
}

func (r *Router) Do(req *Request) {}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	req = req.WithContext(context.WithValue(req.Context(), &rkey, &Request{}))
	r.mux.ServeHTTP(res, req)
}

func (r *Router) request(res http.ResponseWriter, req *http.Request) *Request {
	x := req.Context().Value(&rkey).(*Request)
	x.Reader = req
	x.Writer = res
	return x
}

func (r *Router) mtom(m Middleware) mux.MiddlewareFunc {
	return func(n http.Handler) http.Handler {
		f := EndpointFunc(func(r *Request) { n.ServeHTTP(r.Writer, r.Reader) })
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			m(f).Do(r.request(res, req))
		})
	}
}

var rkey = "covered-request"
