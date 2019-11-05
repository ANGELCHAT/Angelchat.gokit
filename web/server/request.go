package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Request struct {
	writer http.ResponseWriter
	reader *http.Request

	err    error
	body   interface{}
	status int
}

func (r *Request) Header() http.Header { return r.writer.Header() }

func (r *Request) Write(b []byte) (int, error) { return r.writer.Write(b) }

func (r *Request) WriteHeader(statusCode int) {
	r.status = statusCode
	r.writer.WriteHeader(statusCode)
}

func (r *Request) Query(name string, otherwise ...string) string {
	out := r.reader.URL.Query().Get(name)
	if out == "" && len(otherwise) > 0 {
		return otherwise[0]
	}
	return out
}

func (r *Request) Param(name string) string { return mux.Vars(r.reader)[name] }

func (r *Request) Read(body interface{}) *Request {
	return r
}

func (r *Request) Response(body interface{}) *Request { r.body = body; return r }

func (r *Request) Error(err error) *Request { r.err = err; return r }
