package server

import (
	"fmt"

	"encoding/json"

	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/log"
)

type sAction struct {
	Name          string
	Subscriptions []es.SubscriberOptions
}

type subscriber struct {
	es     *es.Service
	sub    *es.Subscriber
	stream chan []es.Message
}

func (s *subscriber) Transmit() <-chan Data {
	out := make(chan Data, 1)
	go func() {
		defer close(out)
		for m := range s.stream {
			log.Debug("es.server.subscriber", "new events")
			var e error
			b, err := json.Marshal(m)
			if err != nil {
				e = fmt.Errorf("sending events to peer %s", err)
				log.Error("es.server.subscriber", e)
			}

			out <- Data{Bytes: b, Error: e}
		}
	}()

	return out
}

func (s *subscriber) Receive(b []byte) error {
	action := sAction{}
	if err := json.Unmarshal(b, &action); err != nil {
		return err
	}

	switch action.Name {
	case "subscription":
		if s.sub != nil {
			if err := s.es.Unsubscribe(s.sub); err != nil {
				log.Error("es.server.subscriber", err)
			}
		}

		// create new event store subscriber, register it and receive
		// messages in s.receiver method.
		sub := es.NewSubscriber(s.receiver, action.Subscriptions...)
		if err := s.es.Subscribe(sub); err != nil {
			log.Error("es.server.subscriber",
				fmt.Errorf("subscribing to ES %s", err))
		}

		s.sub = sub
	}

	return nil

}

func (s *subscriber) Close() error {
	if s.sub != nil {
		if err := s.es.Unsubscribe(s.sub); err != nil {
			log.Error("es.server.subscriber", err)
			return err
		}
	}
	close(s.stream)

	return nil
}

func (s *subscriber) receiver(ms []es.Message) {
	s.stream <- ms
}

func newSubscriber(e *es.Service) TransmitReceiver {
	return &subscriber{
		es:     e,
		stream: make(chan []es.Message, 1),
	}
}
