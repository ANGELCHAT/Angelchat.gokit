package web

import (
	"net/http"

	"github.com/livechat/gokit/web/server"
)

type Server struct{ *server.Service }

func NewServer() *Server { return &Server{server.New()} }

func (s *Server) Run(addr string) error { return http.ListenAndServe(addr, s) }
