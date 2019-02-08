package rest

import (
	"net/http"

	"github.com/sokool/gokit/web/rest/docs"
)

func Run(addr string, s *Server) error { return http.ListenAndServe(addr, s) }

func Documentation(oo ...docs.Option) *docs.Doc {
	return docs.Documentation(oo...)
}
