package server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/livechat/gokit/log"
	"github.com/livechat/gokit/web/server"
)

func TestName(t *testing.T) {
	h := func(r *server.Request) {
		r.Response.Body = map[string]interface{}{
			"Some":      1,
			"CreatedAt": time.Now(),
			"testHereAndThre": map[string]interface{}{
				"a":   2,
				"two": "dwa",
			},
		}
	}

	test := func(label string) server.Middleware { return server.Htom(testA(label)) }
	logger := server.With.Logger(log.Default.Print)
	json := server.With.JSON("snake")
	failed := server.With.Error(nil)
	r := server.New()
	r.
		Prefix("/chat", failed, test("A"), logger, json).
		Prefix("/tags", test("B"), test("C")).
		Handle("/{id}", h, "GET", test("D"))

	s := httptest.NewServer(r)
	c := s.Client()

	o, err := c.Get(s.URL + "/chat/tags/one?q=a")
	if err != nil {
		t.Fatal(err)
	}

	if o.StatusCode != 200 {
		b, _ := ioutil.ReadAll(o.Body)
		t.Fatalf("status 200 is expected, received %s: %s", o.Status, string(b))
	}

	//b, _ := ioutil.ReadAll(o.Body)
	//fmt.Println("out:", string(b))

	o, err = c.Get(s.URL + "/chat/tags/two?q=b")
	if err != nil {
		t.Fatal(err)
	}

	if o.StatusCode != 200 {
		t.Fatalf("status 200 is expected, received: %s", o.Status)
	}

	//b, _ = ioutil.ReadAll(o.Body)
	//fmt.Println("out:", string(b))
}

func testA(label string) func(http.Handler) http.Handler {
	return func(n http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%s: #1\n", label)
			n.ServeHTTP(w, r)
			fmt.Printf("%s: #2\n", label)
		})
	}
}
