package tags

import (
	"fmt"
	"net/http"
	"time"

	"github.com/livechat/gokit/web/client"
)

type Tag struct {
	Name      string `json:"name"`
	Author    string `json:"author"`
	Timestamp int64  `json:"creation_date"`
	In        struct {
		Chats   int `json:"inChats"`
		Tickets int `json:"inTicketsmhm"`
	} `json:"count"`
	Group int `json:"group"`
}

func (t Tag) CreatedAt() time.Time { return time.Unix(t.Timestamp, 0) }

type API struct {
	host string
	http client.Caller
}

func New(url, token string) *API {
	return &API{
		host: url,
		http: client.NewCaller(http.DefaultClient,
			//client.Logging(*log.Default),
			client.Header("X-API-Version", "2"),
			client.ResponseError(nil),
			client.JSONResponse(),
			client.JSONRequest(),
			client.Authorization(token),
		),
	}
}

func (a *API) All() ([]Tag, error) {
	var r []Tag
	return r, a.http.Call(client.Option{
		URL:      fmt.Sprintf("%s/tags", a.host),
		Method:   "GET",
		Response: &r,
	})
}
