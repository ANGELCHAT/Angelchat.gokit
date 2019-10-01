package livechat

import (
	"github.com/livechat/gokit/livechat/crm"
	"github.com/livechat/gokit/livechat/gis"
	"github.com/livechat/gokit/livechat/sso"
	"github.com/livechat/gokit/livechat/tags"
)

type (
	API struct {
		CRM  *crm.API
		Tags *tags.API
		GIS  *gis.API
		SSO  *sso.API
	}

	Logger = func(message string, arguments ...interface{})
)

func New(config ...option) *API {
	o := newOptions(config...)

	sapi, _ := sso.New(o.sso.host)
	return &API{
		CRM:  crm.New(o.crm.host, o.crm.token),
		GIS:  gis.New(o.gis.host, o.gis.token),
		SSO:  sapi,
		Tags: tags.New("", ""),
	}
}
