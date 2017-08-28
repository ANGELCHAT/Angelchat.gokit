package log_test

import (
	"testing"

	"fmt"

	"bytes"

	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
)

type input struct {
	tag string
	msg string
	arg []interface{}
}

type test struct {
	given    input
	expected string
}

func TestLevelWriters(t *testing.T) {

	info := &bytes.Buffer{}
	debug := &bytes.Buffer{}
	err := &bytes.Buffer{}

	logger := log.New(
		log.Levels(info, debug, err),
		log.MessageFormat(false),
	)

	tc := []test{
		{
			input{"log.test", "message 1:%s, 2: %s", []interface{}{"one", "two"}},
			"log.test: message 1:one, 2: two\n"},
		{
			input{"    ", "message", nil},
			"message\n"},
		{
			input{"", "message", nil},
			"message\n"},
		{
			input{"log.test", "", nil},
			"log.test: \n"},
	}

	for _, d := range tc {
		logger.Info(d.given.tag, d.given.msg, d.given.arg...)
		is.Equal(t, d.expected, read(info))

		logger.Debug(d.given.tag, d.given.msg, d.given.arg...)
		is.Equal(t, d.expected, read(debug))

		logger.Error(d.given.tag, fmt.Errorf(d.given.msg, d.given.arg...))
		is.Equal(t, d.expected, read(err))
	}
}

func TestMessageFormat(t *testing.T) {
	output := &bytes.Buffer{}
	prefix := "time.Now() "

	logger := log.New(
		log.Levels(output, output, output),
		log.Formatter(func(m log.Message) string {
			return fmt.Sprintf("%s%s", prefix, m.Text)
		}),
	)

	logger.Info("log.decorator.info", "msg0")
	is.Equal(t, prefix+"msg0\n", read(output))

	logger.Debug("log.decorator.debug", "msg1")
	is.Equal(t, prefix+"msg1\n", read(output))

	logger.Error("log.decorator.error", fmt.Errorf("msg2"))
	is.Equal(t, prefix+"msg2\n", read(output))
}

func TestTagWriter(t *testing.T) {
	var logger *log.Logger
	var tag *bytes.Buffer = &bytes.Buffer{}

	// with empty Info, Debug, Error writers.
	logger = log.New(
		// everything goes into tags writer
		log.Tags(tag, "^.*$"),
		log.MessageFormat(false),
	)

	td := []test{
		{
			input{"", "message0", nil},
			"message0\n"},
		{
			input{"log.infoIn", "message1", nil},
			"log.infoIn: message1\n"},
		{
			input{"some.tag.example", "message2", nil},
			"some.tag.example: message2\n"},
		{
			input{"  a.b.c ", "out %s %s", []interface{}{"a", "b"}},
			"  a.b.c : out a b\n"},
	}

	for _, d := range td {
		logger.Info(d.given.tag, d.given.msg, d.given.arg...)
		is.Equal(t, d.expected, read(tag))

		logger.Debug(d.given.tag, d.given.msg, d.given.arg...)
		is.Equal(t, d.expected, read(tag))

		logger.Error(d.given.tag, fmt.Errorf(d.given.msg, d.given.arg...))
		is.Equal(t, d.expected, read(tag))
	}

	// test one tag writer with multiple tags
	logger = log.New(
		log.Tags(tag, "^log.handler.*$", "^log.test$"),
		log.MessageFormat(false),
	)

	given := []input{
		{"log", "msg0", nil},
		{"log.handler", "msg1", nil},
		{"log.handler.a", "msg2", nil},
		{"log.handler.b", "msg3", nil},
		{"log.handler.b.c", "msg4", nil},
		{"log.test", "msg5", nil},
		{"log.test.a", "msg6", nil},
		{"log.test.a.b", "msg7", nil},
	}

	expects := `log.handler: msg1
log.handler.a: msg2
log.handler.b: msg3
log.handler.b.c: msg4
log.test: msg5
`

	for _, data := range given {
		logger.Info(data.tag, data.msg, data.arg...)
	}
	is.Equal(t, expects, read(tag))

	//
	infoIn, aTagIn, bTagIn := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	logger = log.New(
		log.Levels(infoIn, nil, nil),
		log.Tags(aTagIn, "^a.log$"),
		log.Tags(bTagIn, "^b.log$"),
		log.MessageFormat(false),
	)

	logger.Info("a.log", "msg1")
	logger.Info("b.log", "msg2")

	logger.Debug("a.log", "msg3")
	logger.Debug("b.log", "msg4")

	logger.Error("a.log", fmt.Errorf("msg5"))
	logger.Error("b.log", fmt.Errorf("msg6"))

	aTagOut := `a.log: msg1
a.log: msg3
a.log: msg5
`
	bTagOut := `b.log: msg2
b.log: msg4
b.log: msg6
`
	infoOut := `a.log: msg1
b.log: msg2
`
	is.Equal(t, aTagOut, read(aTagIn))
	is.Equal(t, bTagOut, read(bTagIn))
	is.Equal(t, infoOut, read(infoIn))

}

func read(b *bytes.Buffer) string {
	o := b.String()
	b.Reset()

	return o
}
