package store

import (
	"fmt"
	"sync"

	"github.com/sokool/gokit/log"
)

//type event struct {
//	id        string
//	aggregate string
//	data      []byte
//	kind      string
//	version   uint
//}

type mem struct {
	mu *sync.Mutex

	aggregates map[string]Aggregate
	//events     map[string][]event
	snapshots map[string]Snapshot

	// test helper data
	LastLoadAggregate string
	LastLoadVersion   uint
	LastSaveAggregate Aggregate
	LastSaveEvents    []Event
	calls             []string
}

func (m *mem) Aggregates(kind string, f uint) ([]Aggregate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, "last")
	log.Debug("cqrs.store.last",
		"loading %s aggregates older than %d versions",
		kind, f)
	var o []Aggregate
	for _, a := range m.aggregates {
		var version uint
		if a.Type != kind {
			continue
		}
		s, ok := m.snapshots[a.ID]
		if ok {
			version = s.Version
		}

		is := a.Version - version
		if uint(is) < f {
			continue
		}

		o = append(o, a)
	}

	return o, nil
}

func (m *mem) Make(s Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, "make")
	log.Debug("cqrs.store.make",
		"aggregate %s snapshot in %d version",
		s.AggregateID[24:], s.Version)

	m.snapshots[s.AggregateID] = s
	return nil
}

func (m *mem) Load(id string) (Aggregate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, "load")
	log.Debug("cqrs.store.load",
		"loading aggregate by %s ID", id[24:])
	a, ok := m.aggregates[id]
	if !ok {
		return Aggregate{}, fmt.Errorf("aggregate %s not found", id)
	}

	return a, nil
}

func (m *mem) Snapshot(aggregateID string) (uint, []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, "snapshot")
	log.Debug("cqrs.store.snapshot",
		"loading aggregate %s snapshot",
		aggregateID[24:])
	if s, ok := m.snapshots[aggregateID]; ok {
		return s.Version, s.Data
	}

	return 0, []byte{}
}

//func (m *mem) Save(a Aggregate, es []Event) error {
//	m.mu.Lock()
//	defer m.mu.Unlock()
//
//	m.LastSaveAggregate = a
//	m.LastSaveEvents = es
//
//	m.calls = append(m.calls, "save")
//	log.Debug("cqrs.store.save",
//		"%s with %d events",
//		a.String(), len(es))
//
//	m.aggregates[a.ID] = a
//	for _, e := range es {
//		ev := event{
//			id:        e.ID,
//			aggregate: a.ID,
//			version:   e.Version,
//			data:      e.Data,
//			kind:      e.Type,
//		}
//
//		m.events[a.ID] = append(m.events[a.ID], ev)
//	}
//
//	return nil
//}
//
//func (m *mem) Events(id string, fromVersion uint) ([]Event, error) {
//	m.mu.Lock()
//	defer m.mu.Unlock()
//
//	m.calls = append(m.calls, "events")
//	log.Debug("cqrs.store.events",
//		"loading aggregate %s events from version %d",
//		id[24:], fromVersion)
//
//	m.LastLoadAggregate = id
//	m.LastLoadVersion = fromVersion
//
//	var events []Event
//
//	for i := int(fromVersion); i < len(m.events[id]); i++ {
//		e := m.events[id][i]
//		if fromVersion > 0 {
//			e = m.events[id][i-1]
//		}
//
//		events = append(events, Event{
//			ID:      e.id,
//			Type:    e.kind,
//			Data:    e.data,
//			Version: e.version,
//			Created: time.Time{},
//		})
//		//log.Debug("cqrs.store.events", "%s v.%d event loaded",
//		//	e.kind, e.version)
//	}
//
//	return events, nil
//}

//
// Test Helper functions
//
func (m *mem) AggregatesCount() int {
	return len(m.aggregates)
}

//func (m *mem) AggregatesEventsCount(id string) int {
//
//	es, ok := m.events[id]
//	if !ok {
//		return 0
//	}
//
//	return len(es)
//}

func (m *mem) MethodCalls() []string {
	return m.calls
}

func NewMemoryStorage() *mem {
	return &mem{
		mu:         &sync.Mutex{},
		aggregates: map[string]Aggregate{},
		//events:     map[string][]event{},
		snapshots: map[string]Snapshot{},
		calls:     []string{},
	}
}
