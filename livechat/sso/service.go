package sso

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/livechat/gokit/livechat/sso/clients"
	"github.com/livechat/gokit/livechat/sso/proto"
	webclient "github.com/livechat/gokit/web/client"
	"google.golang.org/grpc"
)

var ErrInsufficientScopes = errors.New("insufficient scopes")

type API struct {
	url    *url.URL
	client proto.SSOAPIClient
	HTTP   *HTTP
}

func New(host string) (*API, error) {
	a := API{}
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	c, err := grpc.Dial(u.Host+":93", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	a.url = u
	a.client = proto.NewSSOAPIClient(c)
	a.HTTP = &HTTP{&a}

	return &a, nil
}

func (s *API) Client(token string) *clients.API {
	endpoint := webclient.NewCaller(webclient.Default,
		webclient.ResponseError(httpError),
		webclient.JSONResponse(),
		webclient.JSONRequest(),
		webclient.Authorization(token),
		//webclient.Logging(log.Default.Print),
	)

	return clients.New(s.url.String(), s.client, endpoint)
}

func httpError(i int, body []byte) error {
	switch i {
	case http.StatusUnauthorized:
		return clients.ErrWrongToken
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("required correct values of: %s", body)
	case http.StatusNotFound:
		return clients.ErrNotFound
	default:
		return fmt.Errorf(string(body))
	}
	return nil
}
