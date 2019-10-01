package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/livechat/gokit/livechat/sso/proto"
	"github.com/livechat/gokit/web/client"
)

var (
	ErrWrongToken = errors.New("wrong or expired authorization token")
	ErrNotFound   = errors.New("client not found")
)

type Client struct {
	ID           string  `json:"client_id,omitempty"`
	Name         string  `json:"name,omitempty"`
	Account      string  `json:"account_id,,omitempty"`
	Organization string  `json:"orgarnization_id,omitempty"`
	Secret       string  `json:"secret,omitempty"`
	RedirectURL  string  `json:"redirect_uri,omitempty"`
	Visibility   string  `json:"visibility,omitempty"`
	Type         string  `json:"type,omitempty"`
	Scopes       []Scope `json:"scopes,omitempty"`
}

func (c Client) String() string {
	b, _ := json.MarshalIndent(c, "", "\t")
	return fmt.Sprintf("%T%s", c, string(b))
}

func (c Client) SecretHash() string {
	size := len(c.Secret)
	switch {
	case size >= 16:
		return fmt.Sprintf("%[1]s****%[2]s", c.Secret[:2], c.Secret[size-4:size])
	case size >= 8:
		return fmt.Sprintf("%[1]s****%[2]s", c.Secret[:1], c.Secret[size-1:size])
	case size > 1:
		return fmt.Sprintf("%[1]s****", c.Secret[:1])
	}

	return ""
}

func (c Client) IsEmpty() bool {
	return len(c.ID) == 0
}

type Scope struct {
	Name        string `json:"scope"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type Stats struct {
	Licenses int `json:"licenses_authorized"`
	Entities int `json:"entities_authorized"`
}

type Info struct {
	Client       string `json:"client_id"`
	Entity       string `json:"entity_id"`
	License      int    `json:"license_id"`
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Expires      int    `json:"expires_in"`
}

func (i Info) HasScope(scope ...string) bool {
	for _, s := range scope {
		if !strings.Contains(i.Scope, s) {
			return false
		}
	}

	return true
}

type API struct {
	url      string
	api      proto.SSOAPIClient
	endpoint client.Caller
}

func New(u string, a proto.SSOAPIClient, e client.Caller) *API {
	return &API{u, a, e}
}

func (s *API) Info() (Info, error) {
	in := client.Option{
		URL:      fmt.Sprintf("%s/info", s.url),
		Response: &Info{}}

	return *in.Response.(*Info), s.endpoint.Call(in)
}

func (s *API) All() ([]Client, error) {
	o := map[string][]Client{}
	r := client.Option{
		URL:      fmt.Sprintf("%s/client", s.url),
		Response: &o}

	if err := s.endpoint.Call(r); err != nil {
		return o["result"], err
	}

	return o["result"], nil
}

func (s *API) One(id string) (Client, error) {
	var c Client
	err := s.endpoint.Call(client.Option{
		URL:      s.url + fmt.Sprintf("/client/%s", id),
		Response: &c,
		Method:   "GET",
	})

	return c, err
}

func (s *API) Stats(id string) (Stats, error) {
	var (
		response Stats
		err      = s.endpoint.Call(client.Option{
			URL:      fmt.Sprintf("%s/client/%s/stats", s.url, id),
			Response: &response,
		})
	)

	return response, err
}

func (s *API) Register(name, typ, secret string, url ...string) (string, uint64, error) {
	var res struct {
		Client  string `json:"client_id"`
		License uint64 `json:"license_id"`
	}

	if len(secret) < 10 {
		return "", 0, fmt.Errorf("passowrd requires at least 10 characters")
	}

	err := s.endpoint.Call(client.Option{
		URL:    s.url + "/client",
		Method: "POST",
		Request: map[string]string{
			"name":         name,
			"redirect_uri": strings.Join(url, ","),
			"type":         typ,
			"secret":       secret,
		},
		Response: &res,
	})

	return res.Client, res.License, err
}

func (s *API) Change(id, name, typ, password string, uri ...string) error {
	return s.change(id, fields{
		"name":         name,
		"type":         typ,
		"secret":       password,
		"redirect_uri": strings.Join(uri, ","),
	})
}

func (s *API) Name(id, n string) error { return s.change(id, fields{"name": n}) }

func (s *API) Type(id, typ string) error { return s.change(id, fields{"type": typ}) }

func (s *API) Secret(id, password string) error {
	return s.change(id, fields{"secret": password})
}

func (s *API) RedirectURI(id string, urls ...string) error {
	return s.change(id, fields{"redirect_uri": strings.Join(urls, ",")})
}

func (s *API) Private(id string) error {
	return s.change(id, fields{"visibility": "private"})
}

func (s *API) Public(id string) error {
	return s.change(id, fields{"visibility": "public"})
}

func (s *API) RevokeAccess(id string) error {
	return s.endpoint.Call(client.Option{
		URL:    fmt.Sprintf("%s/client/%s/scopes/grant-access/license", s.url, id),
		Method: "DELETE",
	})
}

func (s *API) Remove(id string) error {
	return s.endpoint.Call(client.Option{
		URL:      s.url + fmt.Sprintf("/client/%s", id),
		Response: nil,
		Method:   "DELETE",
	})
}

func (s *API) Create(in *Client) error {
	if err := s.endpoint.Call(client.Option{
		URL:      s.url + "/client",
		Method:   "POST",
		Request:  *in,
		Response: in}); err != nil {

		return err
	}

	if err := s.AddScopes(in.ID, in.Scopes...); err != nil {
		if derr := s.Delete(in.ID); derr != nil {
			return derr
		}
		return err
	}

	return nil
}

func (s *API) Store(c *Client) error {
	if c.ID == "" {
		return s.Create(c)
	}

	loaded, err := s.Load(c.ID)
	if err != nil {
		return err
	}

	removeScopes := c.Scopes != nil

	if err := s.endpoint.Call(client.Option{
		URL:      fmt.Sprintf("%s/client/%s", s.url, c.ID),
		Method:   "PUT",
		Request:  *c,
		Response: &c}); err != nil {
		return err
	}

	if !removeScopes {
		return nil
	}

	if err := s.RemoveScopes(loaded.ID, loaded.Scopes...); err != nil {
		return err
	}

	if err := s.AddScopes(loaded.ID, c.Scopes...); err != nil {
		return err
	}

	return nil
}

func (s *API) Update(c *Client) error {
	return s.endpoint.Call(client.Option{
		URL:      fmt.Sprintf("%s/client/%s", s.url, c.ID),
		Method:   "PUT",
		Request:  *c,
		Response: &c})
}

func (s *API) Delete(id string) error {
	r := client.Option{
		URL:      s.url + fmt.Sprintf("/client/%s", id),
		Response: &map[string]string{},
		Method:   "DELETE"}

	return s.endpoint.Call(r)
}

func (s *API) RemoveScopes(id string, scopes ...Scope) error {
	var (
		sentinel sync.WaitGroup
		failed   sync.Map
		errors   []Scope
	)

	for i := range scopes {
		sentinel.Add(1)
		go func(c Scope) {
			defer sentinel.Done()
			if err := s.endpoint.Call(client.Option{
				URL:    fmt.Sprintf("%s/client/%s/scopes/%s", s.url, id, c.Name),
				Method: "DELETE"}); err != nil {
				failed.Store(c, err)
			}
		}(scopes[i])
	}

	sentinel.Wait()

	failed.Range(func(key, value interface{}) bool {
		errors = append(errors, key.(Scope))
		return true
	})

	if len(errors) > 0 {
		return fmt.Errorf("invalid scopes %+v", errors)
	}

	return nil
}

func (s *API) AddScopes(id string, scopes ...Scope) error {
	var (
		sentinel sync.WaitGroup
		failed   sync.Map
		err      []Scope
	)
	for i := range scopes {
		sentinel.Add(1)
		go func(scope Scope) {
			defer sentinel.Done()
			if err := s.endpoint.Call(client.Option{
				URL:     fmt.Sprintf("%s/client/%s/scopes/%s", s.url, id, scope.Name),
				Request: fields{"required": scope.Required},
				Method:  "POST"}); err != nil {
				failed.Store(scope, err)
			}
		}(scopes[i])
	}

	sentinel.Wait()

	failed.Range(func(key, value interface{}) bool {
		err = append(err, key.(Scope))
		return true
	})

	if len(err) > 0 {
		return fmt.Errorf("invalid scopes %+v", err)
	}

	return nil
}

func (s *API) Load(id string) (Client, error) {
	var c Client
	r, err := s.api.GetClient(context.Background(), &proto.IDRequest{Value: id})
	if err != nil {
		return c, err
	}

	c.ID = r.Id
	//c.License = int(r.LicenseID)
	c.Name = r.Name
	c.RedirectURL = r.RedirectUri
	c.Secret = r.SecretHash
	c.Type = r.Type
	c.Visibility = r.Visibility
	c.Scopes = []Scope{}
	for i := range r.Scopes {
		c.Scopes = append(c.Scopes, Scope{
			Name:     r.Scopes[i].Scope,
			Required: r.Scopes[i].Required,
		})
	}

	return c, nil
}

type fields map[string]interface{}

func (s *API) change(id string, f fields) error {
	var res map[string]interface{}
	return s.endpoint.Call(client.Option{
		URL:      fmt.Sprintf("%s/client/%s", s.url, id),
		Method:   "PUT",
		Request:  f,
		Response: &res,
	})
}

//ScopesFromString
func ScopesFromString(scopes string) []Scope {
	ss := make([]Scope, 0)
	scopes = strings.Replace(scopes, " ", "", -1)
	if scopes == "" {
		return ss
	}

	for _, scope := range strings.Split(scopes, ",") {
		p := strings.Split(scope, "[not-required]")
		if len(p) == 1 {
			ss = append(ss, Scope{Name: p[0], Required: true})
			continue
		}

		ss = append(ss, Scope{Name: p[0], Required: false})
	}

	return ss
}

func ScopesToStrings(ss []Scope) []string {
	var o []string
	for _, s := range ss {
		n := s.Name
		//if s.Required {
		//	n += "[not-required]"
		//}
		//
		o = append(o, n)
	}

	return o
}
