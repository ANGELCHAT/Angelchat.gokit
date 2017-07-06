package log

import (
	"io"
	"os"
)

var (
	Debugging bool
	Default   Logger
)

func init() {
	Default = New(os.Stdout, 0)
}

func Info(n, m string, a ...interface{}) {
	Default.Info(n, m, a...)
}

func Debug(n, m string, a ...interface{}) {
	Default.Debug(n, m, a...)
}

func Error(namespace, format string, arguments ...interface{}) {
	Default.Error(namespace, format, arguments...)
}

func Subscribe(w io.Writer, ns ...string) {
	Default.Subscribe(w, ns...)
}
