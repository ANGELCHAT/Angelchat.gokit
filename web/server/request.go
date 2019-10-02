package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Request struct {
	Reader *http.Request
	Writer *writer
}

func (r *Request) Query(name string, otherwise ...string) string {
	out := r.Reader.URL.Query().Get(name)
	if out == "" && len(otherwise) > 0 {
		return otherwise[0]
	}
	return out
}

func (r *Request) Read(body interface{}) *Request {
	return r
}

func (r *Request) Param(name string) string        { return mux.Vars(r.Reader)[name] }
func (r *Request) Write(body interface{}) *Request { r.Writer.body = body; return r }
func (r *Request) Error(err error) *Request        { r.Writer.err = err; ; return r }

type writer struct {
	r      http.ResponseWriter
	err    error
	body   interface{}
	status int
}

func (w *writer) Header() http.Header { return w.r.Header() }

func (w *writer) Write(b []byte) (int, error) { return w.r.Write(b) }

func (w *writer) WriteHeader(statusCode int) { w.status = statusCode; w.r.WriteHeader(statusCode) }
