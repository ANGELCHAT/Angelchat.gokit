package client

import (
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/sokool/gokit/log"
)

type Event struct {
	TopicID   string
	TopicName string
	EventID   string
	EventName string
	Version   uint
	Meta      []byte
	Data      []byte
}

type event struct {
	ID   string
	Name string
	Data []byte
}

type stream struct {
	ID     string
	Name   string
	Meta   []byte
	Events []event
}

type Subscription struct {
	ID          string
	Stream      string
	Events      []string
	FromVersion uint
}

type Client interface {
	NewSubscriber(s ...Subscription) (*Subscriber, error)
	NewWriter(stream string) (*Writer, error)
}

type cli struct {
	r *websocket.Conn
	w *websocket.Conn
}

func (c *cli) NewSubscriber(s ...Subscription) (*Subscriber, error) {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:9999", Path: "/subscribe"}
	read, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	log.Info("es.client.reader", "%s connected...", u.String())

	return &Subscriber{
		conn:          read,
		subscriptions: s,
	}, nil
}

func (c *cli) NewWriter(stream string) (*Writer, error) {
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:9999", Path: "/stream"}
	write, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	log.Info("es.client.writer", "%s connected...", u.String())

	return &Writer{
		stream: stream,
		conn:   write,
	}, nil
}

func NewWebSocket() Client {
	return &cli{}
}
