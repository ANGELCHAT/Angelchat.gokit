package cqrs

import (
	"fmt"

	"time"

	"github.com/sokool/gokit/log"
)

type Store interface {
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
		var sv uint64
		if a.Type != kind {
			continue
		}
		//log.Debug("cqrs.store.last", "%s", a.String())
		s, ok := m.snapshots[a.ID]
		if ok {
			sv = s.Version
		}

		is := a.Version - sv

		if uint(is) < frequency {
			log.Debug("cqrs.store.last",
				"every %d, waiting for %d more events",
				frequency, frequency-uint(is))
			continue
		}

		o = append(o, CQRSAggregate{
			ID:      a.ID,
			Version: sv,
			Type:    a.Type,
		})
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
	// this method should be transactional
	// check if aggregate has not been changed by other request!
	if l, err := m.Load(a.ID); err == nil {
		if (a.Version - uint64(len(es))) != l.Version {
			return fmt.Errorf(
				"%s version missmatch, arrived: %d, expects: %d",
				a.Type, a.Version, l.Version)
		}
	}

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

	//log.Debug("cqrs.store.events",
	//	"aggregate:%s from %d version", id, fromVersion)

	var events []Event

	//if fromVersion > 0 {
	//	for i := int(fromVersion); i <= len(m.events[id]); i++ {
	//		log.Debug("cqrs.store.events.iterator", "version:%d, %d",
	//			i, m.events[id][i-1].version)
	//	}
	//}

	for _, e := range m.events[id] {
		//log.Debug("cqrs.store.events.loading", "%+v", e)

		events = append(events, Event{
			ID:      e.id,
			Type:    e.kind,
			Data:    e.data,
			Version: e.version,
			Created: time.Time{},
		})
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
