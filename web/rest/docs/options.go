package docs

import "github.com/sokool/gokit/web/log"

type Options struct {
	Storage    Endpoints
	HttpPrefix string
	Verbose    bool
}

type Option func(*Options)

func WithEndpoints(r Endpoints) Option {
	return func(o *Options) {
		o.Storage = r
	}
}

func WithVerbose(v bool) Option {
	return func(o *Options) {
		o.Verbose = v
		verbose = v
	}
}

func WithPrefix(p string) Option {
	return func(o *Options) {
		o.HttpPrefix = p
	}
}

func newOptions(ops ...Option) *Options {
	s := &Options{}
	for _, o := range ops {
		o(s)
	}

	if s.Storage == nil {
		s.Storage = newRepository()
	}

	return s

}

var verbose = true

func info(f string, args ...interface{}) {

	log.Default.Info(f, args...)
}

func debug(f string, args ...interface{}) {
	if verbose {
		log.Default.Debug(f, args...)
	}
}
