package web

import (
	"net/http"

	"github.com/livechat/gokit/web/server"
)

type Server struct{ *server.Router }

func NewServer() *Server { return &Server{server.NewRouter()} }

func (s *Server) Run(addr string) error { return http.ListenAndServe(addr, s) }
