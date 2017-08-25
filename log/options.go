package log

import (
	"io"
	"log"
)

type Options struct {
	Info                io.Writer
	Error               io.Writer
	Debug               io.Writer
	Colors              bool
	OutputHandlers      map[string][]io.Writer
	OutputDecoratorFunc func(s string) string
}

type Option func(*Options)

func InfoWriter(w io.Writer) Option {
	return func(o *Options) {
		o.Info = w
	}
}

func ErrorWriter(w io.Writer) Option {
	return func(o *Options) {
		o.Error = w
	}
}

func DebugWriter(w io.Writer) Option {
	return func(o *Options) {
		o.Debug = w
	}
}

func NoColors() Option {
	return func(o *Options) {
		o.Colors = false
	}
}

func OutputDecorator(fn func(string) string) Option {
	return func(o *Options) {
		o.OutputDecoratorFunc = fn
	}
}

func OutputHandler(w io.Writer, tags ...string) Option {
	return func(o *Options) {
		for _, n := range tags {
			o.OutputHandlers[n] = append(o.OutputHandlers[n], w)
		}
	}
}

func newOptions(ops ...Option) *Options {
	s := &Options{
		Colors:         true,
		OutputHandlers: make(map[string][]io.Writer),
	}

	for _, o := range ops {
		o(s)
	}

	var p []string

	if s.Info == nil {
		p = append(p, "InfoWriter")
	}

	if s.Debug == nil {
		p = append(p, "DebugWriter")
	}

	if s.Error == nil {
		p = append(p, "ErrorWriter")
	}

	if len(p) != 0 {
		log.Printf("no %v ware attached\n", p)
	}

	return s

}
