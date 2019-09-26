package server

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/livechat/gokit/web/server/docs"
)

type Request struct {
	Reader   *http.Request
	Writer   http.ResponseWriter
	Response struct {
		Body  interface{}
		Error error
	}
}

func (r *Request) Query(name string, otherwise ...string) string {
	out := r.Reader.URL.Query().Get(name)
	if out == "" && len(otherwise) > 0 {
		return otherwise[0]
	}
	return out
}

func (r *Request) Param(name string) string { return mux.Vars(r.Reader)[name] }

func (r *Request) Return(v interface{}, err error) { r.Response.Body, r.Response.Error = v, err }

type Endpoint interface {
	Do(*Request)
}

type EndpointFunc func(*Request)

func (f EndpointFunc) Do(r *Request) { f(r) }

type Middleware func(Endpoint) Endpoint

func Documentation(oo ...docs.Option) *docs.Doc {
	return docs.Documentation(oo...)
}
