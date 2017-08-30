package cqrs

type HandlerFunc func(Event) error

type Options struct {
	Handlers []HandlerFunc
	Storage  Store
	Name     string
}

type Option func(*Options)

func Storage(s Store) Option {
	return func(o *Options) {
		o.Storage = s
	}
}

func HandleEvent(fn HandlerFunc) Option {
	return func(o *Options) {
		if o.Handlers == nil {
			o.Handlers = []HandlerFunc{}
		}

		o.Handlers = append(o.Handlers, fn)
	}
}

func Name(n string) Option {
	return func(o *Options) {
		o.Name = n
	}
}

func newOptions(ops ...Option) *Options {
	s := &Options{}

	for _, o := range ops {
		o(s)
	}

	if s.Storage == nil {
		s.Storage = newMemStorage()
	}

	return s

}
