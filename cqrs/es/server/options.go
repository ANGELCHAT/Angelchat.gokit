package server

import (
	"fmt"

	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/log"
)

type options struct {
	address    string
	eventStore *es.Service
}

type Option func(*options)

func Address(s string) Option {
	return func(o *options) {
		o.address = s
	}
}

func newOptions(ops ...Option) *options {
	opts := &options{}

	for _, o := range ops {
		o(opts)
	}

	if len(opts.address) == 0 {
		opts.address = "localhost:9999"
	}

	if opts.eventStore == nil {
		var err error
		opts.eventStore, err = es.NewService("/tmp/ws-test")
		if err != nil {
			log.Fatal("es.server", fmt.Errorf("event store %s", err.Error()))
		}
	}

	return opts

}
