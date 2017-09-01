package cqrs

import (
	"fmt"

	"github.com/sokool/gokit/log"
)

type Store interface {
	Save(Aggregate) error
	Load(string) (Aggregate, error)
}

type memStorage struct {
	store map[string]*Aggregate
}

//todo store aggregate{id, name, version, []event}!
func (s *memStorage) Save(a Aggregate) error {
	log.Info("cqrs.mem-store.save", "%s.%dv: #%s + %d new events",
		a.Name, a.Version, a.ID[24:], len(a.Events))

	// no aggregate = add it!
	if _, ok := s.store[a.ID]; !ok {
		s.store[a.ID] = &a
		s.store[a.ID].Version = uint64(len(a.Events))
		for _, e := range a.Events {
			log.Debug("cqrs.mem-store.save.event", e.String())
		}

		return nil
	}

	// nothing has changed!
	if len(a.Events) == 0 {
		return nil
	}

	current := s.store[a.ID]
	if current.Version != a.Version {
		return fmt.Errorf("#%s mismatch versions, got: %d, expected: %d",
			a.ID[24:], a.Version, current.Version)
	}

	for _, e := range a.Events {
		current.Events = append(current.Events, e)
		log.Debug("cqrs.mem-store.save.event", e.String())
		current.Version++
	}

	return nil
}

// todo load aggregate{id, name, version, []event}
func (s *memStorage) Load(id string) (Aggregate, error) {
	a, ok := s.store[id]
	if !ok {
		return *a, fmt.Errorf("#%s not found", id)
	}

	log.Info("cqrs.mem-store.load", "%s.%dv: #%s",
		a.Name, a.Version, a.ID[24:])

	for _, e := range a.Events {
		log.Debug("cqrs.mem-store.load", "event %s", e.String())
	}

	return *a, nil
}

func newMemStorage() Store {
	return &memStorage{
		store: map[string]*Aggregate{},
	}
}
