package cqrs

import (
	"fmt"

	"time"

	"github.com/sokool/gokit/log"
)

type Store interface {
	//SNAPSHOT METHODS

	// Last is calculated by subtracting the last snapshot version
	// from the current version with a where clause that only returned the
	// aggregates with a difference greater than some number. This query
	// will return all of the Aggregates that a snapshot to be created.
	// The snapshotter would then iterate through this list of Aggregates
	// to create the snapshots (if using	multiple snapshotters the
	// competing consumer pattern works well here).
	Last(kind string, vFrequency uint) ([]CQRSAggregate, error)
	Make(s Snapshot) error
	Snapshot(aggregate string) (uint64, []byte)

	//EVENT METHODS
	Load(id string) (CQRSAggregate, error)
	Save(CQRSAggregate, []Event) error
	// load all aggregates and events from given version.
	Events(version uint64, aggregate string) ([]Event, error)
}

type event struct {
	id        string
	aggregate string
	data      []byte
	kind      string
	version   uint64
}

type mem struct {
	aggregates map[string]CQRSAggregate
	events     map[string][]event
	snapshots  map[string]Snapshot

	// test helper data
	LastLoadID      string
	LastLoadVersion uint64
}

func (m *mem) Make(s Snapshot) error {
	m.snapshots[s.AggregateID] = s
	return nil
}

func (m *mem) Last(kind string, frequency uint) ([]CQRSAggregate, error) {
	var o []CQRSAggregate
	for _, a := range m.aggregates {
		var version uint64
		if a.Type != kind {
			continue
		}
		s, ok := m.snapshots[a.ID]
		if ok {
			version = s.Version
		}

		is := a.Version - version
		if uint(is) < frequency {
			continue
		}

		o = append(o, a)
	}

	return o, nil
}

func (m *mem) Load(id string) (CQRSAggregate, error) {
	a, ok := m.aggregates[id]
	if !ok {
		return CQRSAggregate{}, fmt.Errorf("aggregate %s not found", id)
	}

	return a, nil
}

func (m *mem) Snapshot(aggregateID string) (uint64, []byte) {
	if s, ok := m.snapshots[aggregateID]; ok {
		return s.Version, s.Data
	}

	return 0, []byte{}
}

func (m *mem) Save(a CQRSAggregate, es []Event) error {
	m.aggregates[a.ID] = a
	for _, e := range es {
		m.events[a.ID] = append(m.events[a.ID], event{
			id:        e.ID,
			aggregate: a.ID,
			version:   e.Version,
			data:      e.Data,
			kind:      e.Type,
		})
	}

	return nil
}

func (m *mem) Events(fromVersion uint64, id string) ([]Event, error) {
	m.LastLoadID = id
	m.LastLoadVersion = fromVersion

	var events []Event

	for i := int(fromVersion); i < len(m.events[id]); i++ {
		e := m.events[id][i]
		if fromVersion > 0 {
			e = m.events[id][i-1]
		}

		events = append(events, Event{
			ID:      e.id,
			Type:    e.kind,
			Data:    e.data,
			Version: e.version,
			Created: time.Time{},
		})
		log.Debug("cqrs.store.events", "loading event %s v.%d",
			e.kind, e.version)
	}

	return events, nil
}

//
// Test Helper functions
//
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
		aggregates: map[string]CQRSAggregate{},
		events:     map[string][]event{},
		snapshots:  map[string]Snapshot{},
	}
}
