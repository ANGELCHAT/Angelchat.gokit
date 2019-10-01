package server

import "github.com/livechat/gokit/web/server/docs"

func Documentation(oo ...docs.Option) *docs.Doc {
	return docs.Documentation(oo...)
}
