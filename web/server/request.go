package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Request struct {
	Reader   *http.Request
	Writer   http.ResponseWriter
	Response struct {
		Body   interface{}
		Error  error
		Status int
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

func (r *Request) Return(v interface{}, err error) {
	if err != nil {
		r.Response.Error = err
	}

	r.Response.Body = v
}
