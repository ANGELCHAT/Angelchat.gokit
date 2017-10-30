package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sync"
)

type kind int

type Logger struct {
	opt *Options
	mu  sync.Mutex
}

type Message struct {
	Tag   string
	Text  string
	Args  []interface{}
	Kind  kind
	Color string
}

const (
	red    = "\x1b[31;1m%s\x1b[0m"
	green  = "\x1b[32;1m%s\x1b[0m"
	yellow = "\x1b[33;1m%s\x1b[0m"

	info kind = iota
	debug
	err
	tag
)

var Default = New(
	Levels(os.Stdout, nil, os.Stdout),
)

func (l *Logger) Info(tag, msg string, args ...interface{}) {
	l.write(l.opt.Info, Message{
		tag,
		msg,
		args,
		info,
		green},
	)
}

func (l *Logger) Debug(tag, msg string, args ...interface{}) {
	if l.opt.Debug == nil {
		return
	}
	l.write(l.opt.Debug, Message{
		tag,
		msg,
		args,
		debug,
		yellow},
	)

}

func (l *Logger) Error(tag string, e error) {
	l.write(l.opt.Error, Message{
		tag,
		e.Error(),
		nil,
		err,
		red},
	)
}

func (l *Logger) write(w io.Writer, m Message) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if w != nil {
		fmt.Fprintln(w, l.opt.Formatter(m))
	}

	m.Kind = tag
	for t, writers := range l.opt.Tags {
		for _, w := range writers {
			ok, err := regexp.MatchString(t, m.Tag)
			if err != nil {
				log.Printf("output handler tag has %s", err.Error())
			}

			if !ok {
				continue
			}

			fmt.Fprintln(w, l.opt.Formatter(m))
		}
	}
}

func New(os ...Option) *Logger {
	l := &Logger{
		opt: newOptions(os...),
	}

	return l
}

func Info(tag, msg string, args ...interface{}) {
	Default.Info(tag, msg, args...)
}

func Debug(tag, msg string, args ...interface{}) {
	Default.Debug(tag, msg, args...)
}

func Error(tag string, e error) {
	Default.Error(tag, e)
}

func Fatal(tag string, e error) {
	Default.Error(tag, e)
	os.Exit(-1)
}
