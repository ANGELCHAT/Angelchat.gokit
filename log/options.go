package log

import (
	"io"
	"io/ioutil"
	"os"
)

type Options struct {
	Info                io.Writer
	Error               io.Writer
	Debug               io.Writer
	Colors              bool
	Subscribers         map[string]io.Writer
	DecorateMessageFunc func(s string) string
}

type Option func(*Options)

func InfoWriter(l io.Writer) Option {
	return func(o *Options) {
		o.Info = l
	}
}

func ErrorWriter(l io.Writer) Option {
	return func(o *Options) {
		o.Error = l
	}
}

func DebugWriter(l io.Writer) Option {
	return func(o *Options) {
		o.Debug = l
	}
}

func Colors() Option {
	return func(o *Options) {
		o.Colors = true
	}
}

func Decorator(fn func(string) string) Option {
	return func(o *Options) {
		o.DecorateMessageFunc = fn
	}
}

func Listen(w io.Writer, namespaces ...string) Option {
	return func(o *Options) {
		for _, n := range namespaces {
			o.Subscribers[n] = w
		}
	}
}

func newOptions(ops ...Option) *Options {
	s := &Options{
		Info:  os.Stdout,
		Error: os.Stderr,
		Debug: ioutil.Discard,
	}

	for _, o := range ops {
		o(s)
	}

	return s

}
