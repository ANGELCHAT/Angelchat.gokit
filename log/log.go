package log

import (
	"fmt"
	"io"
	"sync"
	"strings"
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

func (l *Log) Info(tag, msg string, args ...interface{}) {
	l.write(l.opt.Info, green, tag, msg, args...)
}

func (l *Log) Debug(tag, msg string, args ...interface{}) {
	l.write(l.opt.Debug, yellow, tag, msg, args...)
}

func (l *Log) Error(tag string, e error) {
	l.write(l.opt.Error, red, tag, e.Error())
}

func (l *Log) write(w io.Writer, color, space, msg string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg = fmt.Sprintf("%s: %s", space, fmt.Sprintf(msg, args...))
	if len(strings.TrimSpace(space)) == 0 {
		msg = msg[len(space)+2:]
	}

	if l.opt.OutputDecoratorFunc != nil {
		msg = l.opt.OutputDecoratorFunc(msg)
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

func Info(tag, msg string, args ...interface{}) {
	Default.Info(tag, msg, args...)
}

func Debug(tag, msg string, args ...interface{}) {
	Default.Debug(tag, msg, args...)
}

func Error(space string, e error) {
	Default.Error(space, e)
}


//func (l *defLog) Subscribe(w io.Writer, ns ...string) Logger {
//	i := log.New(w, "", l.log.Flags())
//	if len(ns) == 0 {
//		l.subscribers["*"] = append(l.subscribers["*"], i)
//		return l
//	}
//
//	for _, n := range ns {
//		l.subscribers[n] = append(l.subscribers[n], i)
//	}
//
//	return l
//}
//
//func (l *defLog) output(n, m string, a ...interface{}) string {
//	m = fmt.Sprintf(m, a...)
//
//	l.notify(n, m)
//
//	if n == "" {
//		return m
//	}
//
//	return n + ": " + m
//}
//
//func (l *defLog) notify(n, m string) {
//	o := fmt.Sprintf("%s: %s", n, m)
//	if n == "" {
//		o = m
//	}
//
//	for p, ws := range l.subscribers {
//		ok, err := regexp.MatchString(p, n)
//		if err != nil {
//			fmt.Println(err)
//			os.Exit(-1)
//		}
//
//		if !ok {
//			continue
//		}
//
//		for _, i := range ws {
//			i.Print(o)
//		}
//
//	}
//
//}
