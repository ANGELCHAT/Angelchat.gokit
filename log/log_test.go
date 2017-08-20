package log_test

import (
	"testing"

	"fmt"

	"github.com/sokool/gokit/log"
)

func TestNew(t *testing.T) {
	tw := newTw()

	l := log.New(
		log.InfoWriter(tw),
	)

	log.Info("", "b")

	tc := map[string]string{
		"log.test.info":          "info testing",
		"some.strange.namespace": "info testing",
		"": "none message",
	}

	for namespace, message := range tc {
		l.Info(namespace, message)
		fmt.Print(tw.Last())
	}

}

type tw struct {
	t *testing.T
	m string
}

func (t *tw) Write(b []byte) (int, error) {
	t.m = string(b)

	return len(b), nil
}

func (t *tw) Last() string {
	return t.m
}

func newTw() *tw {
	return &tw{}
}
