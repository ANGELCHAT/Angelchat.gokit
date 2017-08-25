package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

type Logger struct {
	opt *Options
	mu  sync.Mutex
}

const (
	red    = "\x1b[31;1m%s\x1b[0m"
	green  = "\x1b[32;1m%s\x1b[0m"
	yellow = "\x1b[33;1m%s\x1b[0m"
)

var Default *Logger = New(
	InfoWriter(os.Stdout),
	DebugWriter(os.Stdout),
	ErrorWriter(os.Stdout),
)

func (l *Logger) Info(tag, msg string, args ...interface{}) {
	l.write(l.opt.Info, green, tag, msg, args...)
}

func (l *Logger) Debug(tag, msg string, args ...interface{}) {
	l.write(l.opt.Debug, yellow, tag, msg, args...)
}

func (l *Logger) Error(tag string, e error) {
	l.write(l.opt.Error, red, tag, e.Error())
}

func (l *Logger) write(w io.Writer, color, tag, msg string, args ...interface{}) {
	if w == nil && len(l.opt.OutputHandlers) == 0 {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	msg = fmt.Sprintf("%s: %s", tag, fmt.Sprintf(msg, args...))
	if len(strings.TrimSpace(tag)) == 0 {
		msg = msg[len(tag)+2:]
	}

	if l.opt.OutputDecoratorFunc != nil {
		msg = l.opt.OutputDecoratorFunc(msg)
	}

	if l.opt.OutputHandlers != nil {

		fmt.Fprintln(l.mergeOutputHandlers(tag), msg)
	}

	if w == nil {
		return
	}

	if l.opt.Colors && len(color) != 0 {
		fmt.Fprintln(w, fmt.Sprintf(color, msg))
		return
	}

	fmt.Fprintln(w, msg)
}

func (l *Logger) mergeOutputHandlers(tag string) io.Writer {
	var writers []io.Writer

	for p, ws := range l.opt.OutputHandlers {
		ok, err := regexp.MatchString(p, tag)
		if err != nil {
			log.Printf("output handler tag has %s", err.Error())
		}

		if !ok {
			continue
		}

		writers = append(writers, ws...)

	}

	return io.MultiWriter(writers...)
}

func New(os ...Option) *Logger {
	return &Logger{
		opt: newOptions(os...),
	}
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
