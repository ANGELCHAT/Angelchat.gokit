package log_test

import (
	"testing"

	"fmt"

	"time"

	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
)

type testCase struct {
	tag    string
	msg    string
	args   []interface{}
	output string
}

func TestOutputWriters(t *testing.T) {
	w := newTw()
	l := log.New(
		log.InfoWriter(w),
		log.DebugWriter(w),
		log.ErrorWriter(w),
		log.NoColors(),
	)

	tc := []testCase{
		{
			"log.test",
			"message 1:%s, 2: %s",
			[]interface{}{"one", "two"},
			"log.test: message 1:one, 2: two\n"},
		{
			"    ",
			"message",
			nil,
			"message\n"},
		{
			"",
			"message",
			nil,
			"message\n"},
		{
			"log.test",
			"",
			nil,
			"log.test: \n"},
	}

	for _, c := range tc {
		l.Info(c.tag, c.msg, c.args...)
		is.Equal(t, c.output, w.read())

		l.Debug(c.tag, c.msg, c.args...)
		is.Equal(t, c.output, w.read())

		l.Error(c.tag, fmt.Errorf(c.msg, c.args...))
		is.Equal(t, c.output, w.read())
	}
}

func TestOutputDecorator(t *testing.T) {
	l := log.New(
		log.NoColors(),
		log.OutputDecorator(func(s string) string {
			return time.Now().Format("15:04:05 ") + s
		}))

	l.Info("log.decorator.test", "%s world", "hello")
}

type tw struct {
	t *testing.T
	m string
}

func (t *tw) Write(b []byte) (int, error) {
	t.m = string(b)

	return len(b), nil
}

func (t *tw) read() string {
	return t.m
}

func newTw() *tw {
	return &tw{}
}
