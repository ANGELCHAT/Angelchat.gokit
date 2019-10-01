package gis

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/livechat/gokit/web/client"
)

type elasticSearch struct {
	Cluster string `json:"cluster_id"`
	Index   string `json:"index"`
	Routing string `json:"routing"`
}
type Configuration struct {
	Licence      int           `json:"licenceID"`
	Account      string        `json:"accountID"`
	Agent        string        `json:"operatorID"`
	Organization string        `json:"organizationID"`
	Region       string        `json:"region"`
	Archives     elasticSearch `json:"es"`
	RTD          elasticSearch `json:"es_rtd"`
	Mysql        struct {
		Host     string `json:"host"`
		Database string `json:"db_name"`
		URI      string `json:"uri"`
	} `json:"mysql"`
}

func (l *Configuration) String() string {
	b, _ := json.MarshalIndent(l, "", "\t")
	return string(b)
}

type API struct {
	host, token string
	http        client.Caller
}

func New(host, token string) *API {
	return &API{
		host:  host,
		token: token,
		http: client.NewCaller(http.DefaultClient,
			client.ResponseError(nil),
			client.JSONRequest(),
			client.JSONResponse(),
			//client.Logging(*log.Default),
			client.Authorization(token),
		)}
}

func (a *API) License(number string) (Configuration, error) {
	var r Configuration
	return r, a.http.Call(client.Option{
		URL:      fmt.Sprintf("%s/configuration/licence/%s?password=%s", a.host, number, a.token),
		Response: &r,
	})
}

func (a *API) Agent(mail string) (Configuration, error) {
	var r Configuration
	return r, a.http.Call(client.Option{
		URL:      fmt.Sprintf("%s/configuration/operator/%s?password=%s", a.host, mail, a.token),
		Response: &r,
	})
}
