package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	wlog "github.com/sokool/gokit/web/log"
)

const tag = "publisher"

var log = wlog.New(os.Stdout, tag, true)

type Publisher struct {
	send chan data
}

type data struct {
	url  string
	data interface{}
}

func (d *Publisher) Publish(url string, v interface{}) {
	d.send <- data{url: url, data: v}
}

func (d *Publisher) run() {
	for e := range d.send {
		if len(e.url) == 0 {
			log.Info("%s", fmt.Errorf("no url"))
			continue
		}

		data, err := json.Marshal(e.data)
		if err != nil {
			log.Info("json.Marshal failed due %s", err)
			continue
		}

		res, err := http.Post(e.url, "application/json", bytes.NewReader(data))
		if err != nil {
			log.Info("http.POST failed due %s", err)
			continue
		}

		res.Body.Close()

		if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
			log.Info("%s", fmt.Errorf("received %d status from %s", res.StatusCode, e.url))
			continue
		}

		log.Info("data send to %s", e.url)
	}
}

func NewPublisher() *Publisher {
	s := &Publisher{
		send: make(chan data, 16),
	}

	go s.run()

	return s
}
