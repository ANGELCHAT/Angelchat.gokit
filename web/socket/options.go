package socket

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sokool/gokit/web/log"
)

type options struct {
	shutdown context.Context
	log      logger
}

type option func(*options)

func WithLogger(l log.Logger) option        { return func(o *options) { o.log = l } }
func WithShutdown(c context.Context) option { return func(o *options) { o.shutdown = c } }

func newOptions(oo ...option) (*options, error) {
	o := &options{}
	for i := range oo {
		oo[i](o)
	}

	if o.shutdown == nil {
		o.shutdown = shutdown(o.log.WithTag("shutdown"))
	}

	return o, nil
}

func shutdown(l logger) context.Context {
	termination, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal)
	signal.Notify(stop,
		syscall.SIGTERM,
		os.Interrupt,
		os.Kill,
	)

	go func() {
		l.Debug("wait for SIGTERM, SIGINT, SIGKILL")
		sig := <-stop
		cancel()

		l.Debug("process closed: %+v", sig)

		os.Exit(0)
	}()
	return termination
}
