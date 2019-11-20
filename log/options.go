package log

import (
	"fmt"
	"io"
	"log"
	"log/syslog"
	"os"
	"path"
	"runtime"
	"strings"
	"time"
)

type (
	Tag      struct{ Name, Value string }
	Tagger   func(message []byte, args ...interface{}) Tag
	Renderer func(io.Writer, ...Tag) error
	Option   func(*logger)
)

func construct(configurations ...Option) *logger {
	var c logger
	var err error

	for i := range configurations {
		configurations[i](&c)
	}

	c.taggers = append(c.taggers, func(m []byte, args ...interface{}) Tag {
		return Tag{"message", fmt.Sprintf(string(m), args...)}
	})

	if c.syslog != 0 {
		if c.writer, err = syslog.New(c.syslog, ""); err != nil {
			fallback.Print(err)
		}
	}

	if c.writer == nil {
		c.writer = os.Stdout
	}

	if c.global {
		log.SetFlags(0)
		log.SetOutput(&c)
		standard = &c
	}

	if c.render == nil {
		c.render = writer
	}

	return &c
}

func WithGlobal(b bool) Option { return func(o *logger) { o.global = b } }

func WithName(n string) Option {
	return func(o *logger) {
		o.taggers = append(o.taggers, func([]byte, ...interface{}) Tag {
			return Tag{"name", n}
		})
	}
}

func WithVerbose(v bool) Option {
	return func(o *logger) {
		if !v {
			//for i := range o.taggers {
			//	o.taggers[i].
			//}
			//delete(o.taggers, "verbose")
			return
		}
		o.taggers = append(o.taggers, func([]byte, ...interface{}) Tag {
			return Tag{Name: "verbose"}
		})
	}
}

func WithLevels(colors bool) Option {
	color := func(level string) string {
		if colors {
			switch level {
			case "INF":
				return "\x1b[32;1mINF\x1b[0m" // green
			case "ERR":
				return "\x1b[31;1mERR\x1b[0m"
			case "DBG":
				return "\x1b[33;1mDBG\x1b[0m"
			case "EMG":
				return "\x1b[34;1mEMG\x1b[0m"
			case "WRN":
				return "\x1b[36;1mWRN\x1b[0m"
			}
		}

		return level
	}

	return func(o *logger) {
		o.taggers = append(o.taggers, func(message []byte, args ...interface{}) Tag {
			var size = len(message)
			var level string

			for i := range args {
				if _, ok := args[i].(error); !ok {
					continue
				}

				return Tag{"level", color("ERR")}
			}

			if size == 0 {
				return Tag{"level", "INF"}
			}

			switch message[size-1:][0] {
			case '.':
				level = "DBG"
			case '?':
				level = "WRN"
			case ';':
				level = "EMG"
			case '!':
				level = "ERR"
			default:
				level = "INF"
			}

			if level != "INF" {
				message[size-1] = ' '
			}

			return Tag{"level", color(level)}
		})
	}
}

func WithFilename(depth int) Option {
	// TODO automatically find where log method has been called.
	return func(o *logger) {
		if depth <= 0 {
			return
		}

		o.taggers = append(o.taggers, func([]byte, ...interface{}) Tag {
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

			return Tag{"file", fmt.Sprintf("%s:%d", path.Base(file), line)}
		})
	}
}

func WithTime(format string) Option {
	return func(o *logger) {
		o.taggers = append(o.taggers, func([]byte, ...interface{}) Tag {
			return Tag{"time", time.Now().Format(format)}
		})
	}
}

func WithRenderer(r Renderer) Option {
	return func(o *logger) { o.render = r }
}

func WithTagger(t ...Tagger) Option {
	return func(o *logger) { o.taggers = append(o.taggers, t...) }
}

func WithWriter(w io.Writer) Option { return func(o *logger) { o.writer = w } }

func WithSyslog(p syslog.Priority) Option { return func(o *logger) { o.syslog = p } }

func writer(w io.Writer, tags ...Tag) error {
	//for i := range tags {
	//	fmt.Print(tags[i])
	//}
	//fmt.Println()

	var (
		message string
		verbose bool
	)

	for _, t := range tags {
		if t.Name == "" {
			continue
		}

		switch t.Name {
		case "verbose":
			verbose = true
			continue
		}

		if t.Value == "" {
			continue
		}

		message += fmt.Sprintf("%s ", t.Value)
	}

	if verbose {
		//fmt.Println("verbo!")
	}

	message = strings.TrimSpace(message) + "\n"

	//var _, verbose = t["verbose"]
	//var level = t["level"]
	//
	//if !verbose && level != "DBG" && level != "ERR" {
	//	return
	//}
	//message := fmt.Sprintf("%s %s%s %s %s %s\n", t["name"], t["time"], t["a"], t["file"], t["level"], t["message"])

	_, err := w.Write([]byte(message))

	return err
}
