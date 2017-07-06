package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

type Logger interface {
	Info(namespace, message string, arguments ...interface{})
	Debug(namespace, message string, arguments ...interface{})
	Error(namespace, message string, arguments ...interface{})
	Subscribe(w io.Writer, ns ...string) Logger
}

type logger struct {
	log         *log.Logger
	subscribers map[string][]*log.Logger
}

const (
	red = "\x1b[31;1m%s\x1b[0m"
	//green = "\x1b[38;5;28m%s\x1b[0m"
	green  = "\x1b[32;1m%s\x1b[0m"
	yellow = "\x1b[33;1m%s\x1b[0m"
)

func (l *logger) Error(n, m string, a ...interface{}) {
	m = fmt.Sprintf(red, m)
	n = fmt.Sprintf(red, n)

	l.log.Println(l.output(n, m, a...))
}

func (l *logger) Info(n, m string, a ...interface{}) {
	m = fmt.Sprintf(green, m)
	n = fmt.Sprintf(green, n)

	l.log.Println(l.output(n, m, a...))

}

func (l *logger) Debug(n, m string, a ...interface{}) {
	if !Debugging {
		return
	}

	m = fmt.Sprintf(yellow, m)
	n = fmt.Sprintf(yellow, n)

	l.log.Println(l.output(n, m, a...))
}

func (l *logger) Subscribe(w io.Writer, ns ...string) Logger {
	i := log.New(w, "", l.log.Flags())
	if len(ns) == 0 {
		l.subscribers["*"] = append(l.subscribers["*"], i)
		return l
	}

	for _, n := range ns {
		l.subscribers[n] = append(l.subscribers[n], i)
	}

	return l
}

func (l *logger) output(n, m string, a ...interface{}) string {
	m = fmt.Sprintf(m, a...)

	l.notify(n, m)

	if n == "" {
		return m
	}

	return n + ": " + m
}

func (l *logger) notify(n, m string) {
	o := fmt.Sprintf("%s: %s", n, m)
	if n == "" {
		o = m
	}

	for p, ws := range l.subscribers {
		ok, err := regexp.MatchString(p, n)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}

		if !ok {
			continue
		}

		for _, i := range ws {
			i.Print(o)
		}

	}

}

func New(w io.Writer, f int) Logger {
	return &logger{
		log:         log.New(w, "", f),
		subscribers: make(map[string][]*log.Logger),
	}
}
