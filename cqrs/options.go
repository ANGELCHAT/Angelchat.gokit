package cqrs

import "time"

// todo: custom logger implementation
// todo: custom id generator - separate for events and aggregator?
//		 do I need id for event since I have uint Version?
// todo: every loaded aggregate is kept in memory(cache), only generated events are stored
// 		 it is a form of caching, memoization?
// todo rebuild aggregate based on manually given version and/or date?
// todo consider snapshoting on save instead seperate process

// for external use ie. another aggregate
type HandlerFunc func(CQRSAggregate, []Event, []interface{})

type Options struct {
	Handlers      []HandlerFunc
	Store         Store
	Name          string
	SnapEpoch     uint
	SnapFrequency time.Duration
	Cache         bool
}

type Option func(*Options)

func WithStorage(s Store) Option {
	return func(o *Options) {
		o.Store = s
	}
}

//func Logger() Option {
//	return func(o *Options) {
//
//	}
//}
//
//func IdentityGenerator() Option {
//	return func(o *Options) {
//
//	}
//}

func WithCache() Option {
	return func(o *Options) {
		o.Cache = true
	}
}

func WithSnapshot(epoch uint, frequency time.Duration) Option {
	return func(o *Options) {
		o.SnapEpoch = epoch
		o.SnapFrequency = frequency
	}
}

func EventHandler(fn HandlerFunc) Option {
	return func(o *Options) {
		if o.Handlers == nil {
			o.Handlers = []HandlerFunc{}
		}

		o.Handlers = append(o.Handlers, fn)
	}
}

//func MongoStorage(url, session, collection string) Option {
//	return func(o *Options) {
//		// initialize databases
//		if err := mongo.RegisterSession(session, url); err != nil {
//			log.Error("cqrs.mongo", err)
//			os.Exit(-1)
//		}
//
//		db, err := mongo.Session(session)
//		if err != nil {
//			log.Error("cqrs.mongo", err)
//			os.Exit(-1)
//		}
//
//		o.WithStorage = mongoStore(db, collection)
//	}
//}

func newOptions(ops ...Option) *Options {
	s := &Options{}

	for _, o := range ops {
		o(s)
	}

	if s.Store == nil {
		s.Store = NewMemoryStorage()
	}

	return s

}
