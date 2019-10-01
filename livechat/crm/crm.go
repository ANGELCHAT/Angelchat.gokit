package crm

import (
	"net/http"

	"github.com/livechat/gokit/web/client"
)

type API struct {
	Customers    Customers
	Integrations integrations
}

func New(host, token string) *API {
	c := client.NewCaller(http.DefaultClient,
		client.ResponseError(nil),
		client.JSONRequest(),
		client.JSONResponse(),
		//client.Logging(*log.Default),
		client.Authorization(token),
	)

	return &API{
		Customers:    Customers{http: c, host: host},
		Integrations: integrations{http: c, url: host},
	}
}
