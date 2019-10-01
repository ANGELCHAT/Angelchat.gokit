package server

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

type Service struct{ mux *mux.Router }

func New() *Service { return &Service{mux: mux.NewRouter()} }

func (r *Service) Prefix(prefix string, ms ...Middleware) *Service {
	pr := r.mux.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		pr.Use(Mtoh(m))
	}

	return &Service{mux: pr}
}

func (r *Service) Handle(path string, e EndpointFunc, method string, ms ...Middleware) *Service {
	var h Endpoint = e
	for i := len(ms) - 1; i >= 0; i-- {
		h = ms[i](h)
	}

	f := func(res http.ResponseWriter, req *http.Request) { h.Do(getRequest(res, req)) }
	rh := r.mux.Handle(path, http.HandlerFunc(f))
	rh.Methods(method)

	return r
}

//TODO
func (r *Service) HandleFunc(path string, e http.HandlerFunc, method string, ms ...Middleware) *Service {
	//var h http.Handler = e
	//for i := len(ms) - 1; i >= 0; i-- {
	//	h = ms[i](h)
	//}
	//
	//rh := r.mux.Handle(path, h)
	//rh.Methods(method)

	return r
}

func (r *Service) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(res, setRequest(res, req))
}

func (r *Service) Do(req *Request) {}

func getRequest(res http.ResponseWriter, req *http.Request) *Request {
	return req.Context().Value(&rkey).(*Request)
}

func setRequest(res http.ResponseWriter, req *http.Request) *http.Request {
	r := &Request{}
	r.Reader = req.WithContext(context.WithValue(req.Context(), &rkey, r))
	r.Writer = &writer{r: res}
	return r.Reader
}

var rkey = "covered-request"
