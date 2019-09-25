package publisher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/livechat/gokit/log"
)

const tag = "publisher"

var logs = log.Default.WithTag(tag).Print

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
			logs("%s", fmt.Errorf("no url"))
			continue
		}

		data, err := json.Marshal(e.data)
		if err != nil {
			logs("json.Marshal failed due %s", err)
			continue
		}

		res, err := http.Post(e.url, "application/json", bytes.NewReader(data))
		if err != nil {
			logs("http.POST failed due %s", err)
			continue
		}

		res.Body.Close()

		if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
			logs("%s", fmt.Errorf("received %d status from %s", res.StatusCode, e.url))
			continue
		}

		logs("data send to %s", e.url)
	}
}

func NewPublisher() *Publisher {
	s := &Publisher{
		send: make(chan data, 16),
	}

	go s.run()

	return s
}
