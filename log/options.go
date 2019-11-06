package log

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"path"
	"runtime"
	"time"
)

type options struct {
	writer io.Writer
	syslog syslog.Priority
	//printer  Printer
	override bool

	//todo
	//  use templates such as {TIME-FORMAT}, {PREFIX}, {FILE-NAME}{FILE-LINE}...
	decorators []Decorator
}

type Decorator func(message string, args ...interface{}) string

type Option func(*options)

func WithWriter(w io.Writer) Option { return func(o *options) { o.writer = w } }

func WithSyslog(p syslog.Priority) Option { return func(o *options) { o.syslog = p } }

//func WithPrinter(p Printer) Option  { return func(o *options) { o.printer = p } }
func WithOverride(b bool) Option { return func(o *options) { o.override = b } }

func WithTag(name string) Option {
	return func(o *options) {
		o.decorators = append(o.decorators, func(m string, _ ...interface{}) string {
			return fmt.Sprintf("%s %s", name, m)
		})
	}
}

func WithLevels(verbose bool) Option {
	return func(o *options) {
		o.decorators = append(o.decorators, func(m string, args ...interface{}) string {
			s := len(m)
			if s < 1 {
				return m
			}

			var level string

			switch m[s-1 : s] {
			case ".":
				level = "INF"

			case "?":
				level = "WRN"

			case ";":
				level = "EMG"

			case "!":
				level = "ERR"

			default:
				level = "DBG"
				for i := range args {
					if _, ok := args[i].(error); ok {
						level = "ERR"
						break
					}
				}
				s++
			}

			if !verbose {
				if level == "INF" || level == "ERR" {
					return fmt.Sprintf("[%s] %s", level, m[:s-1])
				}
				return ""
			}

			return fmt.Sprintf("[%s] %s", level, m[:s-1])
		})
	}
}

// TODO automatically find where log method has been called.
func WithFilename(depth int) Option {
	return func(o *options) {
		o.decorators = append(o.decorators, func(m string, _ ...interface{}) string {
			if depth <= 0 {
				return m
			}

			//t := true
			//for n := 0; t; n++ {
			//	_, file, line, ok := runtime.Caller(n)
			//	if !ok {
			//		t = false
			//		break
			//	}
			//	fmt.Println(n, ok, file, line)
			//}

			//https://stackoverflow.com/questions/35212985/is-it-possible-get-information-about-caller-function-in-golang
			_, file, line, _ := runtime.Caller(depth)

			return fmt.Sprintf("%s %s:%d", m, path.Base(file), line)
		})
	}
}

func WithTime(format string) Option {
	return func(o *options) {
		o.decorators = append(o.decorators, func(m string, _ ...interface{}) string {
			return fmt.Sprintf("%s %s", time.Now().Format(format), m)
		})
	}
}

func WithDecorator(d Decorator) Option {
	return func(o *options) {
		o.decorators = append(o.decorators, d)
	}
}

func construct(w io.Writer, configurations ...Option) *options {
	o := &options{writer: w, override: true}

	for i := range configurations {
		configurations[i](o)
	}

	if o.syslog != 0 {
		o.writer, _ = syslog.New(o.syslog, "")
	}

	if o.writer == nil {
		o.writer = os.Stdout
	}

	if o.override {
		log.SetOutput(o.writer)
	}

	return o
}
