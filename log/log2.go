package log

import (
	"fmt"
	"io"
	"sync"
)

type Log struct {
	opt *Options
	mu  sync.Mutex
}

const (
	red    = "\x1b[31;1m%s\x1b[0m"
	green  = "\x1b[32;1m%s\x1b[0m"
	yellow = "\x1b[33;1m%s\x1b[0m"
)

var Default *Log = New()

func (l *Log) Info(space, msg string, args ...interface{}) {
	l.write(l.opt.Info, green, space, msg, args...)
}

func (l *Log) Debug(space, msg string, args ...interface{}) {
	l.write(l.opt.Debug, yellow, space, msg, args...)
}

func (l *Log) Error(space string, e error) {
	l.write(l.opt.Error, red, space, e.Error())
}

func (l *Log) write(w io.Writer, color, space, msg string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg = fmt.Sprintf("%s: %s", space, fmt.Sprintf(msg, args...))
	if len(space) == 0 {
		msg = msg[2:]
	}

	if l.opt.DecorateMessageFunc != nil {
		msg = l.opt.DecorateMessageFunc(msg)
	}

	if l.opt.Colors {
		fmt.Fprintln(w, fmt.Sprintf(color, msg))
		return
	}

	fmt.Fprintln(w, msg)
}

func New(os ...Option) *Log {
	return &Log{
		opt: newOptions(os...),
	}
}

func Info(space, msg string, args ...interface{}) {
	Default.Info(space, msg, args...)
}

func Debug(space, msg string, args ...interface{}) {
	Default.Debug(space, msg, args...)
}

func Error(space string, e error) {
	Default.Error(space, e)
}
