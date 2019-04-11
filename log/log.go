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
	w         io.Writer
	verbose   bool
	colors    bool
	timestamp bool
	tag       string
}

func New(w io.Writer, tag string, verbose bool) *Logger {
	_, timestamp := w.(*os.File)
	if tag != "" {
		tag = fmt.Sprintf("%s", tag)
	}

	return &Logger{
		w:         w,
		verbose:   verbose,
		colors:    w == os.Stdout,
		tag:       tag,
		timestamp: timestamp,
	}
}

func (l *Logger) Write(p []byte) (n int, err error) {
	s := string(p)
	if p := strings.Index(s, "[DEBUG] "); p != -1 {
		l.Print("DBG " + s[p+8:])
		return
	}

	if p := strings.Index(s, "[ERROR] "); p != -1 {
		l.Print("ERR " + s[p+8:])
		return
	}

	if p := strings.Index(s, "[INFO] "); p != -1 {
		l.Print("INF " + s[p+7:])
		return
	}

	l.Print("INF " + s)
	return
}

func (l *Logger) Print(format string, a ...interface{}) {
	s := strings.Split(format, " ")
	typ := strings.ToUpper(s[0])

	if typ != "INF" && typ != "DBG" && typ != "ERR" {
		format = "INF " + format
		s[0] = "INF"
		typ = "INF"
	}

	format = strings.TrimSpace(strings.Replace(format, s[0], "", 1))

	if len(a) >= 1 {
		if _, ok := a[0].(error); ok {
			typ = "ERR"
		}
	}

	if typ == "DBG" && !l.verbose {
		return
	}

	// syslog support
	if w, ok := l.w.(*syslog.Writer); ok {
		m := fmt.Sprintf("%s %s %s", typ, l.tag, fmt.Sprintf(format, a...))

		switch typ {
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
		switch typ {
		case "INF":
			color = "\x1b[32;1m%s\x1b[0m" // green

		case "ERR":
			color = "\x1b[31;1m%s\x1b[0m" // red

		case "DBG":
			color = "\x1b[33;1m%s\x1b[0m" // yellow
		}
	}

	format = fmt.Sprintf(format, a...)
	format = strings.TrimSuffix(format, "\n")
	typ = fmt.Sprintf(color, typ)
	x := l.tag
	if l.tag != "" {
		x = fmt.Sprintf("[\x1b[36;1m%s\x1b[0m] ", l.tag)
	}
	n := time.Now().Format("2006/01/02 15:04:05.000")
	l.w.Write([]byte(fmt.Sprintf("%s [%s] %s%s\n", n, typ, x, format)))
}

func (l *Logger) WithTag(name string) *Logger { return New(l.w, name, l.verbose) }
