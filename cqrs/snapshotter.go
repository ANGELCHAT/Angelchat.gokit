package cqrs

import (
	"time"

	"github.com/sokool/gokit/log"
)

// Making a snapshot algorythym:
// todo: combine 1,2 into one?
// 1. Load all Aggregates of given type where
// 	  aggregate.version - last_snap.version > frequency
// 2. Load last snapshot and restore it on Aggregate.
// 3. Load all the Events from snap.version and process them on Aggregate
// 4. Take snapshot of Aggregate
// 5. Save snapshot of Aggregate with version = Aggregate.version
type snapshotter struct {
	// todo move them out of snapshotter? used only in run method
	frequency uint
	kind      string
	factory   FactoryFunc

	events     *events
	serializer *serializer
	snapshot   structure
}

func (s *snapshotter) run(frequency time.Duration) {
	log.Info("cqrs.snapshot", "starting %s snapshotter every %s and every %d version",
		s.kind, frequency, s.frequency)

	for range time.NewTicker(frequency).C {
		log.Info("cqrs.snapshot", "running...")
		//1. load aggregates... snapshooter
		aggregates, err := s.events.store.Last(s.kind, s.frequency)
		if err != nil {
			log.Error("cqrs.snapshot.load-last", err)
			return
		}

		for _, a := range aggregates {
			aggregate := s.factory(a.ID, a.Version)

			if err := s.take(aggregate); err != nil {
				log.Error("cqrs.snapshot", err)
			}
		}
	}

	log.Info("cqrs.snapshot", "%s finished", s.kind)

}

func (s *snapshotter) take(a *Aggregate) error {
	//2. load last snap and restore aggregate with given snapshot data
	version, data := s.events.store.Snapshot(a.ID)
	if len(data) > 0 {
		snapshot, err := s.serializer.Unmarshal(s.snapshot.Name, data)
		if err != nil {
			return err
		}

		if err := a.RestoreSnapshot(snapshot); err != nil {
			return err
		}
	}

	// 3. Load all the Events from snap.version and process them on Aggregate
	num, err := s.events.load(a, version)
	if err != nil {
		return err
	}

	// 4. Take snapshot of Aggregate
	data, err = s.serializer.Marshal(s.snapshot.Name, a.TakeSnapshot())
	if err != nil {
		return err
	}

	// 5. Save snapshot of Aggregate with version = Aggregate.version
	v := Snapshot{a.ID, data, a.Version}
	if err = s.events.store.Make(v); err != nil {
		return err
	}

	log.Info(
		"cqrs.snapshot", "%s.#%s.v%d taken, rebuilded "+
			"from .v%d, with %d processed events",
		a.Name, a.ID[24:], a.Version, version, num)

	return nil
}

func (s *snapshotter) restore(a *Aggregate) (uint, error) {
	version, data := s.events.store.Snapshot(a.ID)

	// we have snapshot, restore it!
	if len(data) > 0 {
		snapshot, err := s.serializer.Unmarshal(s.snapshot.Name, data)
		if err != nil {
			return 0, err
		}

		if err := a.RestoreSnapshot(snapshot); err != nil {
			return 0, err
		}
	}

	log.Info("cqrs.snapshot",
		"%s.#%s.v%d restored from .v%d",
		a.Name, a.ID[24:], a.Version, version)

	return version, nil
}

func newSnapshotter(frequency uint, e *events, f FactoryFunc) *snapshotter {
	a := f("", 0)
	sStruct := a.TakeSnapshot()
	return &snapshotter{
		frequency:  frequency,
		events:     e,
		factory:    f,
		kind:       a.Name,
		snapshot:   newStructure(sStruct),
		serializer: newSerializer(sStruct),
	}
}
