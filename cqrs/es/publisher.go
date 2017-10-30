package es

import (
	"fmt"

	"github.com/sokool/gokit/log"
	"sync"
)

type MessageHandler func([]Message)

type Message struct {
	EventID   string
	EventName string
	TopicID   string
	TopicName string
	Version   uint
	Meta      []byte
	Data      []byte
}

type Publisher struct {
	subs map[*Subscriber]bool
	sync.Mutex
}

func (p *Publisher) send(z Provider, es []Event) {
	var ms []Message
	for _, event := range es {
		ms = append(ms, Message{
			EventID:   event.ID,
			EventName: event.Type,
			TopicID:   z.ID,
			TopicName: z.Type,
			Version:   z.Version,
			Meta:      event.Meta,
			Data:      event.Data,
		})
	}
	p.Lock()
	for s := range p.subs {
		s.publish(ms)
	}
	p.Unlock()
}

func (p *Publisher) Subscribe(s *Subscriber) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.subs[s]; ok {
		return fmt.Errorf("already subscribed")
	}

	p.subs[s] = true

	log.Debug("es.publisher", "subscriber registered (total:%d)",
		len(p.subs))

	return nil
}

func (p *Publisher) Unsubscribe(s *Subscriber) error {
	p.Lock()
	defer p.Unlock()

	if _, ok := p.subs[s]; !ok {
		return fmt.Errorf("not subscribed")
	}

	delete(p.subs, s)

	log.Debug("es.publisher", "subscriber deregistered (total:%d)",
		len(p.subs))

	return nil
}

// none - read everything
// stream names - read from given stream names
// stream ids - read from given stream ID's
// events name - read all with given event names
//
type SubscriberOptions struct {
	Stream      string
	Events      []string
	FromVersion uint
}

func (s *SubscriberOptions) Ok(stream string, event string) bool {
	if s.Stream != stream {
		return false
	}

	if len(s.Events) == 0 {
		return true
	}

	for _, e := range s.Events {
		if e == event {
			return true
		}
	}

	return false

}

func (s *SubscriberOptions) IsZero() bool {
	if len(s.Stream) == 0 {
		return true
	}

	return false
}

type Subscriber struct {
	//stream   chan []Message
	options  []SubscriberOptions
	receiver MessageHandler
}

func (s *Subscriber) publish(in []Message) {
	var out []Message

	if len(s.options) == 0 {
		s.options = append(s.options, SubscriberOptions{})
	}

	for _, m := range in {
		for _, o := range s.options {
			if !o.IsZero() && !o.Ok(m.TopicName, m.EventName) {
				continue
			}
			out = append(out, m)
		}
	}

	s.receiver(out)
}

func newPublisher() *Publisher {
	return &Publisher{
		subs: make(map[*Subscriber]bool, 1),
	}
}

func NewSubscriber(h MessageHandler, o ...SubscriberOptions) *Subscriber {
	return &Subscriber{
		receiver: h,
		options:  o,
	}
}
