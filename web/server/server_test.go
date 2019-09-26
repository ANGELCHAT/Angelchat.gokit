package server_test

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/livechat/gokit/log"
	"github.com/livechat/gokit/web/server"
)

func TestName(t *testing.T) {
	h := func(r *server.Request) {
		r.Return(map[string]interface{}{
			"Some":      1,
			"CreatedAt": time.Now(),
			"testHereAndThre": map[string]interface{}{
				"a":   2,
				"two": "dwa",
			},
		}, nil)
	}

	logger := server.With.Logger(log.Default.Print)
	json := server.With.JSON("")

	r := server.NewRouter()
	r.
		Prefix("/chat", logger, json).
		Prefix("/tags").
		Handle("/{id}", h, "GET")

	s := httptest.NewServer(r)
	c := s.Client()

	o, err := c.Get(s.URL + "/chat/tags/one?q=a")
	if err != nil {
		t.Fatal(err)
	}

	if o.StatusCode != 200 {
		t.Fatalf("status 200 is expected")
	}

	b, _ := ioutil.ReadAll(o.Body)
	fmt.Println("out:", string(b))

	o, err = c.Get(s.URL + "/chat/tags/two?q=b")
	if err != nil {
		t.Fatal(err)
	}

	if o.StatusCode != 200 {
		t.Fatalf("status 200 is expected")
	}

	b, _ = ioutil.ReadAll(o.Body)
	fmt.Println("out:", string(b))
}
