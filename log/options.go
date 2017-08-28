package log

import (
	"fmt"
	"io"
	"strings"
)

type FormatterFunc func(Message) string

type Options struct {
	Info        io.Writer
	Error       io.Writer
	Debug       io.Writer
	Formatter   FormatterFunc
	TagHandlers []func(Message)
	Tags        map[string][]io.Writer
}

type Option func(*Options)

func Levels(info, debug, error io.Writer) Option {
	return func(o *Options) {
		o.Info, o.Debug, o.Error = info, debug, error
	}
}

func Tags(w io.Writer, tags ...string) Option {
	return func(o *Options) {
		for _, t := range tags {
			o.Tags[t] = append(o.Tags[t], w)
		}
	}
}

func Formatter(f FormatterFunc) Option {
	return func(o *Options) {
		o.Formatter = f
	}
}

func MessageFormat(colors bool) Option {
	return func(o *Options) {
		o.Formatter = func(m Message) string {
			message := fmt.Sprintf("%s: %s", m.Tag, fmt.Sprintf(m.Text, m.Args...))
			if len(strings.TrimSpace(m.Tag)) == 0 {
				message = message[len(m.Tag)+2:]
			}

			if colors && len(m.Color) != 0 {
				message = fmt.Sprintf(m.Color, message)
			}

			return message
		}
	}
}

func newOptions(ops ...Option) *Options {
	s := &Options{
		Tags:        make(map[string][]io.Writer),
		TagHandlers: []func(Message){},
	}

	for _, o := range ops {
		o(s)
	}

	if s.Formatter == nil {
		MessageFormat(true)(s)
	}

	return s

}
