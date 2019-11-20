package log

import (
	"io"
	"log"
	"log/syslog"
	"os"
)

var (
	standard Logger
	fallback *log.Logger
)

func init() {
	standard = New(
		WithWriter(os.Stdout),
		WithLevels(false),
		WithTime("2006-01-02 15:04:05.000000"))

	fallback = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)
}

type Logger interface {
	Print(message string, arguments ...interface{})
	Write(message []byte) (n int, err error)
	New(configuration ...Option) Logger
}

func New(configuration ...Option) Logger        { return construct(configuration...) }
func Print(message string, args ...interface{}) { standard.Print(message, args...) }
func Write(p []byte) (n int, err error)         { return standard.Write(p) }

type logger struct {
	render  Renderer
	writer  io.Writer
	taggers []Tagger

	syslog syslog.Priority
	global bool
}

func (l *logger) Print(message string, arguments ...interface{}) {
	tags := l.decorate([]byte(message), arguments...)
	if err := l.render(l.writer, tags...); err != nil {
		fallback.Print(err)
	}
}

func (l *logger) Write(message []byte) (n int, err error) {
	if err = l.render(l.writer, l.decorate(message)...); err != nil {
		return 0, err
	}

	return len(message), nil
}

func (l *logger) New(configuration ...Option) Logger {
	panic("implement me")
}

func (l *logger) decorate(message []byte, args ...interface{}) []Tag {
	var tags []Tag
	for _, tagger := range l.taggers {
		if tagger != nil {
			tags = append(tags, tagger(message, args...))
		}
	}

	return tags
}

func Syslog() { //todo

}
