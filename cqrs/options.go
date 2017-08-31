package cqrs

import (
	"os"

	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/mongo"
)

// for external use ie. another aggregate
type HandlerFunc func(Identity, Event, interface{}) error

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

func EventHandler(fn HandlerFunc) Option {
	return func(o *Options) {
		if o.Handlers == nil {
			o.Handlers = []HandlerFunc{}
		}

		o.Handlers = append(o.Handlers, fn)
	}
}

func MongoStorage(url, session, collection string) Option {
	return func(o *Options) {
		// initialize databases
		if err := mongo.RegisterSession(session, url); err != nil {
			log.Error("cqrs.mongo", err)
			os.Exit(-1)
		}

		db, err := mongo.Session(session)
		if err != nil {
			log.Error("cqrs.mongo", err)
			os.Exit(-1)
		}

		o.Storage = mongoStore(db, collection)
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
