package cqrs

import (
	"fmt"

	"time"
)

type Store interface {
	Aggregate(id string) (Aggregate, error)
	Save(Aggregate, []Event) error
	Load(string) (Aggregate, []Event, error)
}

type aggregate struct {
	id      string
	kind    string
	version uint64
}

type event struct {
	id        string
	aggregate string
	data      []byte
	kind      string
	version   uint64
}

type snapshot struct {
	aggregate string
	data      []byte
	version   uint64
}

type mem struct {
	aggregates map[string]aggregate
	events     map[string][]event
	snapshots  map[string][]snapshot
}

func (m *mem) Aggregate(id string) (Aggregate, error) {
	a, ok := m.aggregates[id]
	if !ok {
		return Aggregate{}, fmt.Errorf("aggregate %s not found", id)
	}

	return Aggregate{
		ID:      a.id,
		Type:    a.kind,
		Version: a.version,
	}, nil
}

func (m *mem) Save(a Aggregate, es []Event) error {
	// this method should be transactional
	// check if aggregate has not been changed by other request!
	if l, err := m.Aggregate(a.ID); err == nil {
		if (a.Version - uint64(len(es))) != l.Version {
			return fmt.Errorf(
				"%s version missmatch, arrived: %d, expects: %d",
				a.Type, a.Version, l.Version)
		}
	}

	m.aggregates[a.ID] = aggregate{
		id:      a.ID,
		version: a.Version,
		kind:    a.Type,
	}

	for _, e := range es {
		m.events[a.ID] = append(m.events[a.ID], event{
			id:        e.ID.String(),
			aggregate: a.ID,
			version:   e.Version,
			data:      e.Data,
			kind:      e.Type,
		})
	}

	return nil
}

func (m *mem) Load(id string) (Aggregate, []Event, error) {
	out, err := m.Aggregate(id)
	var events []Event

	if err != nil {
		return out, events, err
	}

	for _, e := range m.events[id] {
		events = append(events, Event{
			ID:      Identity(e.id),
			Type:    e.kind,
			Data:    e.data,
			Version: e.version,
			Created: time.Time{},
		})
	}

	return out, events, nil
}

func (m *mem) AggregatesCount() int {
	return len(m.aggregates)
}

func (m *mem) AggregatesEventsCount(id string) int {
	es, ok := m.events[id]
	if !ok {
		return 0
	}

	return len(es)
}

func NewMemoryStorage() *mem {
	return &mem{
		aggregates: map[string]aggregate{},
		events:     map[string][]event{},
		snapshots:  map[string][]snapshot{},
	}
}
