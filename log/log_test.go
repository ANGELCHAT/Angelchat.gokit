package log_test

import (
	"testing"

	"fmt"

	"bytes"

	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
)

type testCase struct {
	tag      string
	msg      string
	args     []interface{}
	expected string
}

func TestOutputWriting(t *testing.T) {
	infoOut := &bytes.Buffer{}
	debugOut := &bytes.Buffer{}
	errorOut := &bytes.Buffer{}

	logger := log.New(
		log.NoColors(),
		log.InfoWriter(infoOut),
		log.DebugWriter(debugOut),
		log.ErrorWriter(errorOut),
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
		logger.Info(c.tag, c.msg, c.args...)
		is.Equal(t, c.expected, read(infoOut))

		logger.Debug(c.tag, c.msg, c.args...)
		is.Equal(t, c.expected, read(debugOut))

		logger.Error(c.tag, fmt.Errorf(c.msg, c.args...))
		is.Equal(t, c.expected, read(errorOut))
	}
}

func TestOutputDecorator(t *testing.T) {
	output := &bytes.Buffer{}
	prefix := "time.Now() "

	logger := log.New(
		log.NoColors(),
		log.InfoWriter(output),
		log.DebugWriter(output),
		log.ErrorWriter(output),
		log.OutputDecorator(func(s string) string {
			return fmt.Sprintf("%s%s", prefix, s)
		}))

	logger.Info("log.decorator.info", "msg0")
	is.Equal(t, prefix+"log.decorator.info: msg0\n", read(output))

	logger.Debug("log.decorator.debug", "msg1")
	is.Equal(t, prefix+"log.decorator.debug: msg1\n", read(output))

	logger.Error("log.decorator.error", fmt.Errorf("msg2"))
	is.Equal(t, prefix+"log.decorator.error: msg2\n", read(output))
}

func TestOutputHandler(t *testing.T) {
	var logger *log.Logger
	output := &bytes.Buffer{}
	outputHandler := &bytes.Buffer{}

	// with empty Info, Debug, Error writers.
	logger = log.New(
		log.OutputHandler(outputHandler, ".*"),
	)

	// info
	logger.Info("", "message0")
	is.Equal(t, "message0\n", read(outputHandler))

	logger.Info("log.info", "message1")
	is.Equal(t, "log.info: message1\n", read(outputHandler))

	logger.Info("some.tag.example", "message2")
	is.Equal(t, "some.tag.example: message2\n", read(outputHandler))

	// debug
	logger.Debug("", "message0")
	is.Equal(t, "message0\n", read(outputHandler))

	logger.Debug("log.info", "message1")
	is.Equal(t, "log.info: message1\n", read(outputHandler))

	logger.Debug("some.tag.example", "message2")
	is.Equal(t, "some.tag.example: message2\n", read(outputHandler))

	// error
	logger.Error("", fmt.Errorf("message0"))
	is.Equal(t, "message0\n", read(outputHandler))

	logger.Error("log.info", fmt.Errorf("message1"))
	is.Equal(t, "log.info: message1\n", read(outputHandler))

	logger.Error("some.tag.example", fmt.Errorf("message2"))
	is.Equal(t, "some.tag.example: message2\n", read(outputHandler))

	logger = log.New(
		log.InfoWriter(output),
		log.DebugWriter(output),
		log.ErrorWriter(output),
		log.OutputHandler(outputHandler, ".*"),
	)

	// info
	logger.Info("", "message0")
	is.Equal(t, "message0\n", read(outputHandler))

	logger.Info("log.info", "message1")
	is.Equal(t, "log.info: message1\n", read(outputHandler))

	logger.Info("some.tag.example", "message2")
	is.Equal(t, "some.tag.example: message2\n", read(outputHandler))

	// debug
	logger.Debug("", "message0")
	is.Equal(t, "message0\n", read(outputHandler))

	logger.Debug("log.info", "message1")
	is.Equal(t, "log.info: message1\n", read(outputHandler))

	logger.Debug("some.tag.example", "message2")
	is.Equal(t, "some.tag.example: message2\n", read(outputHandler))

	// error
	logger.Error("", fmt.Errorf("message0"))
	is.Equal(t, "message0\n", read(outputHandler))

	logger.Error("log.info", fmt.Errorf("message1"))
	is.Equal(t, "log.info: message1\n", read(outputHandler))

	logger.Error("some.tag.example", fmt.Errorf("message2"))
	is.Equal(t, "some.tag.example: message2\n", read(outputHandler))

	h1Tag1 := "^log.handler.*$"
	h1Tag2 := "^log.test$"
	o := `log.handler: msg1
log.handler.a: msg2
log.handler.b: msg3
log.handler.b.c: msg4
log.test: msg5
`

	h1Out := &bytes.Buffer{}
	logger = log.New(
		log.OutputHandler(h1Out, h1Tag1, h1Tag2),
	)

	logger.Info("log", "msg0")
	logger.Info("log.handler", "msg1")
	logger.Info("log.handler.a", "msg2")
	logger.Info("log.handler.b", "msg3")
	logger.Info("log.handler.b.c", "msg4")
	logger.Info("log.test", "msg5")
	logger.Info("log.test.a", "msg6")
	logger.Info("log.test.a.b", "msg7")

	is.Equal(t, o, read(h1Out))

	h1 := &bytes.Buffer{}
	h2 := &bytes.Buffer{}
	h3 := &bytes.Buffer{}

	logger = log.New(
		log.NoColors(),
		log.InfoWriter(h3),
		log.OutputHandler(h1, "^a.log$"),
		log.OutputHandler(h2, "^b.log$"),
	)

	logger.Info("a.log", "msg1")
	logger.Error("a.log", fmt.Errorf("msg2"))
	logger.Debug("a.log", "msg2")

	logger.Info("b.log", "msg3")
	logger.Error("b.log", fmt.Errorf("msg4"))
	logger.Debug("b.log", "msg5")

	o1 := `a.log: msg1
a.log: msg2
a.log: msg2
`
	is.Equal(t, o1, read(h1))

	o2 := `b.log: msg3
b.log: msg4
b.log: msg5
`
	is.Equal(t, o2, read(h2))

	o3 := `a.log: msg1
b.log: msg3
`
	is.Equal(t, o3, read(h3))

}

func TestNoDebugOutput(t *testing.T) {
	out := &bytes.Buffer{}

	log.New().Debug("tag", "message")
	is.True(t, read(out) == "", "empty output expected")

	log.New(log.DebugWriter(out)).Debug("tag", "message")
	is.True(t, read(out) != "", "debug output expected")

}

func read(b *bytes.Buffer) string {
	o := b.String()
	b.Reset()

	return o
}
