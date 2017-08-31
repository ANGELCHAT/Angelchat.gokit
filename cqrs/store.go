package cqrs

import "github.com/sokool/gokit/log"

type Store interface {
	Save(Identity, []Event) error
	Load(Identity) ([]Event, error)
}

type memStorage struct {
	events map[Identity][]Event
}

//todo aggregate name,
func (s *memStorage) Save(aggregate Identity, rs []Event) error {
	log.Debug("\ncqrs.store.save", string(aggregate))
	for _, e := range rs {
		log.Debug("cqrs.store.save.event", e.String())
	}

	s.events[aggregate] = append(s.events[aggregate], rs...)

	return nil
}

func (s *memStorage) Load(aggregate Identity) ([]Event, error) {
	log.Debug("\ncqrs.store.load", string(aggregate))
	for _, e := range s.events[aggregate] {
		log.Debug("cqrs.store.load.event", e.String())
	}

	return s.events[aggregate], nil
}

func newMemStorage() Store {
	return &memStorage{
		events: map[Identity][]Event{},
	}
}
