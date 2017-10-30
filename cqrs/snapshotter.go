package cqrs

import (
	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/cqrs/snapshotter"
	"github.com/sokool/gokit/log"
)

type snaps struct {
	snapshotter *snapshotter.Repository
	eventStore  *es.EventStore
}

//func (s *snaps) run(frequency time.Duration) {
//	log.Info("cqrs.snapshot", "starting %s snapshotter every %s and every %d version",
//		s.kind, frequency, s.frequency)
//
//	for range time.NewTicker(frequency).C {
//		log.Info("cqrs.snapshot", "running...")
//		//1. load aggregates... snapshooter
//		aggregates, err := s.snapshotter.Persistence.Aggregates(s.kind, s.frequency)
//		if err != nil {
//			log.Error("cqrs.snapshot.load-last", err)
//			return
//		}
//
//		for _, a := range aggregates {
//			aggregate := s.factory(a.ID, a.Version)
//
//			if err := s.take(aggregate); err != nil {
//				log.Error("cqrs.snapshot", err)
//			}
//		}
//	}
//
//	log.Info("cqrs.snapshot", "%s finished", s.kind)
//
//}

func (s *snaps) take(a *Aggregate) error {
	if err := s.restore(a); err != nil {
		return err
	}

	// 4. Take snapshot of Aggregate and save it
	if err := s.snapshotter.Take(a.ID, a.Version, a.TakeSnapshot()); err != nil {
		return err
	}

	return nil
}

func (s *snaps) restore(a *Aggregate) error {
	// load last snapshot
	snapshot, last, err := s.snapshotter.Restore(a.ID)
	if err != nil {
		return err
	}

	// we have snapshot, restore it!
	if last > 0 {
		if err := a.RestoreSnapshot(snapshot); err != nil {
			return err
		}
	}

	// load all the Events from restored version and process them on Aggregate
	events, version, err := s.eventStore.Load(a.ID, last)
	if err != nil {
		return err
	}
	a.Version = version
	a.events = events
	a.apply()
	a.events = []interface{}{}

	log.Info("cqrs.snapshot.restore", "%s.#%s.v%d from .v%d",
		a.Name, a.ID[24:], a.Version, last)

	return nil
}

func newSnapshotter(e *es.EventStore, r *snapshotter.Repository) *snaps {
	return &snaps{
		snapshotter: r,
		eventStore:  e,
	}
}
