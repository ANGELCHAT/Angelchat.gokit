package log

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"strings"
	"time"
)

var Default = New(os.Stdout, "", true)

type Logger struct {
	output    io.Writer
	verbose   bool
	colors    bool
	timestamp bool
	tag       string
}

func New(w io.Writer, tag string, verbose bool) Logger {
	_, timestamp := w.(*os.File)
	if tag != "" {
		tag = fmt.Sprintf("%s", tag)
	}

	return Logger{
		output:    w,
		verbose:   verbose,
		colors:    w == os.Stdout,
		tag:       tag,
		timestamp: timestamp,
	}
}

func (l Logger) Info(f string, args ...interface{}) { l.write("INF", f, args...) }

func (l Logger) Debug(f string, args ...interface{}) { l.write("DBG", f, args...) }

func (l Logger) WithTag(s string) Logger { return New(l.output, s, l.verbose) }

func (l Logger) Write(p []byte) (n int, err error) {
	s := string(p)
	if p := strings.Index(s, "[DEBUG] "); p != -1 {
		l.Debug(s[p+8:])
		return
	}

	if p := strings.Index(s, "[ERROR] "); p != -1 {
		l.Info("%s", fmt.Errorf(s[p+8:]))
		return
	}

	if p := strings.Index(s, "[INFO] "); p != -1 {
		l.Info(s[p+7:])
		return
	}

	l.Info(s)
	return
}

func (l Logger) write(t string, m string, args ...interface{}) {
	if len(args) >= 1 {
		if _, ok := args[0].(error); ok {
			t = "ERR"
		}
	}

	if t == "DBG" && !l.verbose {
		return
	}

	// syslog support
	if w, ok := l.output.(*syslog.Writer); ok {
		m := fmt.Sprintf("%s %s %s", t, l.tag, fmt.Sprintf(m, args...))

		switch t {
		case "INF":
			w.Info(m)
		case "ERR":
			w.Err(m)
		case "DBG":
			w.Debug(m)
		}

		return
	}

	color := "%s"
	if l.colors {
		switch t {
		case "INF":
			color = "\x1b[32;1m%s\x1b[0m" // green

		case "ERR":
			color = "\x1b[31;1m%s\x1b[0m" // red

		case "DBG":
			color = "\x1b[33;1m%s\x1b[0m" // yellow
		}
	}

	m = fmt.Sprintf(m, args...)
	m = strings.TrimSuffix(m, "\n")
	t = fmt.Sprintf(color, t)
	if l.tag != "" {
		l.tag = fmt.Sprintf("[\x1b[36;1m%s\x1b[0m] ", l.tag)
	}
	n := time.Now().Format("2006/01/02 15:04:05.000")

	l.output.Write([]byte(fmt.Sprintf("%s [%s] %s%s\n", n, t, l.tag, m)))
}
