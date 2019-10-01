package sso_test

import (
	"testing"

	"github.com/livechat/gokit/livechat/sso"
	"github.com/livechat/gokit/livechat/sso/clients"
	"github.com/livechat/gokit/test/is"
)

var (
	// this need to be filled up before running intergration tests
	service *sso.API
	token   = "Bearer xxx"
)

func init() {
	var err error
	service, err = sso.New("http://sso.fra02.sl.labs")
	if err != nil {
		panic(err)
	}
}

func TestClientInfo(t *testing.T) {
	i, err := service.Client(token).Info()

	is.Ok(t, err)
	is.True(t, i.License != 0, "expects license")
	is.True(t, i.Entity != "", "expects entity")
	is.True(t, i.TokenType != "", "expects token type")
	is.True(t, i.Client != "", "expects client ")
	is.True(t, i.AccessToken != "", "expects access token")
	is.True(t, i.Scope != "", "expects scopes ")
	is.True(t, i.Expires != 0, "expects expires")
}

func TestNewClient(t *testing.T) {
	c := service.Client(token)

	Name := "my app name"
	Secret := "my-cool-password"
	RedirectURL := "http://onet.pl"
	Type := "server_side_app"
	Scopes := clients.ScopesFromString("agents_read")

	cm := clients.Client{
		Name:        Name,
		Secret:      Secret,
		RedirectURL: RedirectURL,
		Type:        Type,
		Scopes:      Scopes}

	is.Ok(t, c.Create(&cm))

	is.Equal(t, Name, cm.Name)
	is.Equal(t, Secret, cm.Secret)
	is.Equal(t, RedirectURL, cm.RedirectURL)
	is.Equal(t, Type, cm.Type)
	is.Equal(t, Scopes[0].Name, cm.Scopes[0].Name)
	is.Equal(t, Scopes[0].Required, cm.Scopes[0].Required)
	is.True(t, cm.ID != "", "id expected")
	is.True(t, cm.Visibility == "private", "private visibility expected")
	//is.True(t, cm.License != 0, "license expected")
	is.Ok(t, c.Delete(cm.ID))
}

func TestStoreClient(t *testing.T) {
	c := service.Client(token)

	created := clients.Client{
		Name:        "app-name",
		Secret:      "secret-pass",
		RedirectURL: "http://redirect-url",
		Visibility:  "private",
		Type:        "server_side_app",
		Scopes: []clients.Scope{
			{Name: "agents_read", Required: true}}}

	is.Ok(t, c.Create(&created))
	defer c.Delete(created.ID)

	modified := clients.Client{
		ID:          created.ID,
		Name:        "xxx-app",
		RedirectURL: "http://blbla",
		Type:        "javascript_app",
		Visibility:  "public",
		Scopes: []clients.Scope{
			{Name: "tickets_read", Required: false}}}

	is.Ok(t, c.Store(&modified))

	loaded, err := c.Load(modified.ID)
	is.Ok(t, err)

	is.Equal(t, modified.ID, loaded.ID)
	is.Equal(t, modified.Name, loaded.Name)
	is.Equal(t, modified.Type, loaded.Type)
	is.Equal(t, modified.RedirectURL, loaded.RedirectURL)
	is.Equal(t, modified.Visibility, loaded.Visibility)
	is.Equal(t, modified.Scopes[0].Name, loaded.Scopes[0].Name)
	is.Equal(t, modified.Scopes[0].Required, loaded.Scopes[0].Required)
}

//func TestStoreClient2(t *testing.T) {
//	c := service.Client(token)
//	created := clients.Client{
//		Name:        "app-name",
//		Secret:      "secret-pass",
//		RedirectURL: "http://redirect-url",
//		Visibility:  "private",
//		Type:        "server_side_app",
//		Scopes:      clients.ScopesFromString("agents_read,agents_write")}
//
//	is.Ok(t, c.Store(&created))
//	defer c.Delete(created.ID)
//
//	created.Visibility = "public"
//	is.Ok(t, c.Store(&created))
//
//	visibility, err := c.Load(created.ID)
//	is.Ok(t, err)
//
//	is.Equal(t, len(created.Scopes), len(visibility.Scopes))
//	is.Equal(t, created.Scopes, visibility.Scopes)
//
//	c.RemoveScopes(created.ID)
//	visibility2 := &clients.Client{
//		ID:         created.ID,
//		Visibility: "private",
//		Scopes:     clients.ScopesFromString("")} // this removes scopes
//
//	is.Ok(t, c.Store(visibility2))
//	is.Ok(t, c.Load(visibility2))
//
//	is.Equal(t, 0, len(visibility2.Scopes))
//
//}

func TestRevoke(t *testing.T) {
	c := service.Client(token)
	created := clients.Client{
		Name:        "app-name",
		Secret:      "secret-pass",
		RedirectURL: "http://redirect-url",
		Visibility:  "public",
		Type:        "server_side_app",
		Scopes:      clients.ScopesFromString("agents_read,agents_write")}

	is.Ok(t, c.Store(&created))
	is.Ok(t, c.RevokeAccess(created.ID))
	is.Ok(t, c.Delete(created.ID))
}

func TestScopesFromString(t *testing.T) {
	s := clients.ScopesFromString("test")
	is.Equal(t, "test", s[0].Name)
	is.Equal(t, true, s[0].Required)

	s = clients.ScopesFromString("test[not-required]")
	is.Equal(t, "test", s[0].Name)
	is.Equal(t, false, s[0].Required)

	s = clients.ScopesFromString("test[]")
	is.Equal(t, "test[]", s[0].Name)
	is.Equal(t, true, s[0].Required)

	s = clients.ScopesFromString("test[fake]")
	is.Equal(t, "test[fake]", s[0].Name)
	is.Equal(t, true, s[0].Required)

	s = clients.ScopesFromString("test[fake][not-required]")
	is.Equal(t, "test[fake]", s[0].Name)
	is.Equal(t, false, s[0].Required)

	s = clients.ScopesFromString("   ")
	is.Equal(t, 0, len(s))
}

func TestClientSecretHash(t *testing.T) {
	is.Equal(t, "so****ters", clients.Client{Secret: "somePassLongerEqual16Characters"}.SecretHash())
	is.Equal(t, "s****s", clients.Client{Secret: "somePass"}.SecretHash())
	is.Equal(t, "z****r", clients.Client{Secret: "zomeMoreChar"}.SecretHash())
	is.Equal(t, "s****", clients.Client{Secret: "seven7m"}.SecretHash())
	is.Equal(t, "t****", clients.Client{Secret: "tw"}.SecretHash())
	is.Equal(t, "", clients.Client{Secret: "x"}.SecretHash())
	is.Equal(t, "", clients.Client{Secret: ""}.SecretHash())
}
