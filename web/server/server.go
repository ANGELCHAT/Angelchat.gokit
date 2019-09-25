package server

import (
	"net/http"

	"github.com/livechat/gokit/web/server/docs"
)

func Run(addr string, r *Router) error { return http.ListenAndServe(addr, r) }

func Documentation(oo ...docs.Option) *docs.Doc {
	return docs.Documentation(oo...)
}
