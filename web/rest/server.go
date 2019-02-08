package rest

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	Router *mux.Router
}

func NewServer() *Server {
	return &Server{
		Router: mux.NewRouter(),
	}
}

func (r *Server) Prefix(prefix string, ms ...Middleware) *Server {
	pr := r.Router.PathPrefix(prefix).Name(prefix + "prefix").Subrouter()

	for _, m := range ms {
		pr.Use(mux.MiddlewareFunc(m))
	}

	return &Server{Router: pr}
}

func (r *Server) Handle(path string, h http.HandlerFunc, method string, ms ...Middleware) *Server {
	handler := http.Handler(h)
	for i := len(ms) - 1; i >= 0; i-- {
		handler = ms[i](handler)
	}

	rh := r.Router.Handle(path, handler)
	rh.Methods(method)

	return r
}

func (r *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(res, req)
}
