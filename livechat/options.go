package livechat

type service struct{ host, token string }

type options struct {
	gis, sso, crm service
	log           Logger
}

type option func(*options)

func WithCRM(h, t string) option { return func(o *options) { o.crm = service{h, t} } }
func WithSSO(h, t string) option { return func(o *options) { o.sso = service{h, t} } }
func WithGIS(h, t string) option { return func(o *options) { o.gis = service{h, t} } }

// WithLogger
func WithLogger(l Logger) option { return func(o *options) { o.log = l } }

func newOptions(oo ...option) *options {
	o := &options{
		log: func(string, ...interface{}) {},
	}

	for i := range oo {
		oo[i](o)
	}

	return o
}
