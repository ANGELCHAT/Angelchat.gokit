package sso

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/livechat/gokit/livechat/sso/clients"
)

type HTTP struct{ api *API }

// Authorize is a http server middleware - it tries to determine if Authorization
// header exists in http.Request and if has two factors (TokenType, TokenValue) ie
// Authorization: Bearer fra-a:u-DatiFVRIF3W0VITmNfoA. Token is used to connect
// with SSO service and fetch Info about SSO Token. Additionally scopes might
// be matched against SSO Entity.
func (a *HTTP) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token = r.Header.Get("Authorization")
		var info clients.Info
		var err error

		if token != "" {
			if info, err = a.api.Client(token).Info(); err == clients.ErrWrongToken {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			} else if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		r.Header.Set("sso-token-type", info.TokenType)
		r.Header.Set("sso-access-token", info.AccessToken)
		r.Header.Set("sso-refresh-token", info.RefreshToken)
		r.Header.Set("sso-license", strconv.Itoa(info.License))
		r.Header.Set("sso-client", info.Client)
		r.Header.Set("sso-entity", info.Entity)
		r.Header.Set("sso-expires", strconv.Itoa(info.Expires))
		r.Header.Set("sso-scopes", info.Scope)

		h.ServeHTTP(w, r)
	})
}

func (a *HTTP) Scopes(scopes ...string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("sso-entity") == "" {
				http.Error(w, "no user found", http.StatusForbidden)
				return
			}

			if !a.HasScope(r, scopes...) {
				http.Error(w, ErrInsufficientScopes.Error(), http.StatusForbidden)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func (a *HTTP) HasScope(r *http.Request, scopes ...string) bool {
	for _, s := range scopes {
		if !strings.Contains(r.Header.Get("sso-scopes"), s) {
			return false
		}
	}

	return true
}

func (a *HTTP) IsAuthenticatedAs(r *http.Request, licenses ...string) bool {
	license := r.Header.Get("sso-license")

	if len(licenses) == 0 {
		return true
	}

	for i := range licenses {
		if licenses[i] == license {
			return true
		}
	}

	return false

}

// AuthenticatedAs checks if account is one of email address. If no email address
// are used then it checks if account exists at all. Mail can be partial, without
// valid email address.
func (a *HTTP) AuthenticatedAs(licenses ...string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var license string

			if license = r.Header.Get("sso-license"); license == "" {
				http.Error(w, "no account found", http.StatusForbidden)
				return
			}

			if a.IsAuthenticatedAs(r, licenses...) {
				h.ServeHTTP(w, r)
				return
			}

			http.Error(w, "account not allowed", http.StatusForbidden)
		})
	}
}
