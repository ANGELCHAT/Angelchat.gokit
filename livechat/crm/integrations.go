package crm

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/livechat/gokit/web/client"
)

type integrations struct {
	url  string
	http client.Caller
}

type Integration struct {
	ID   string `json:"integration_id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (i *integrations) Activate(application string, auth string) error {
	return i.withOauth(auth).Call(client.Option{
		URL:    fmt.Sprintf("%s/integrations?id=%s", i.url, application),
		Method: http.MethodPost,
	})
}

func (i *integrations) Deactivate(application string, token string) error {
	return i.withOauth(token).Call(client.Option{
		URL:    fmt.Sprintf("%s/integrations/%s", i.url, application),
		Method: http.MethodDelete,
	})
}

func (i *integrations) All() ([]Integration, error) {
	var response struct{ Integrations []Integration }

	return response.Integrations, i.http.Call(client.Option{
		URL:      fmt.Sprintf("%s/crm/integrations", i.url),
		Method:   http.MethodGet,
		Response: &response,
	})
}

func (i *integrations) withOauth(token string) client.Caller {
	region := "dal"
	if strings.Contains(token, "fra:") {
		region = "fra"
	}

	err := func(httpStatus int, body []byte) error {
		if !bytes.Contains(body, []byte("Integration already exists!")) {
			return fmt.Errorf("http(%d) with %s", httpStatus, string(body))
		}
		return nil
	}

	return client.NewCaller(http.DefaultClient,
		client.ResponseError(err),
		client.Header("x-api-version", "2"),
		client.Header("x-region", region),
		client.Authorization(token),
		//client.Logging(log.Default),
	)

}
