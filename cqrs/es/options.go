package es

type Options struct {
	//storage  storage
	Listener []func(Event)
}

type Option func(*Options)

//func WithStore(s storage) Option {
//	return func(o *Options) {
//		o.storage = s
//	}
//}

func WithListener(s func(Event)) Option {
	return func(o *Options) {
		o.Listener = append(o.Listener, s)
	}
}

func newOptions(ops ...Option) *Options {
	s := &Options{}

	for _, o := range ops {
		o(s)
	}

	//if s.storage == nil {
	//	s.storage = NewMemStorage()
	//}

	return s

}
