package log

import (
	"io"
	"os"
)

type Options struct {
	Info                io.Writer
	Error               io.Writer
	Debug               io.Writer
	Colors              bool
	Subscribers         map[string]io.Writer
	OutputDecoratorFunc func(s string) string
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

//func Listen(w io.Writer, namespaces ...string) Option {
//	return func(o *Options) {
//		for _, n := range namespaces {
//			o.Subscribers[n] = w
//		}
//	}
//}

func newOptions(ops ...Option) *Options {
	s := &Options{
		Info:   os.Stdout,
		Error:  os.Stderr,
		Debug:  os.Stdout,
		Colors: true,
	}

	for _, o := range ops {
		o(s)
	}

	return s

}
