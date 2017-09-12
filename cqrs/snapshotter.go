package cqrs

import "github.com/sokool/gokit/log"

type snapshotter struct {
	frequency  uint
	kind       string
	events     *events
	factory    func() Aggregate
	serializer *serializer
	snapshot   structure
}

// Making a snapshot algorythym:
// todo: combine 1,2 into one?
// 1. Load all Aggregates of given type where
// 	  aggregate.version - last_snap.version > frequency
// 2. Load last snapshot and restore it on Aggregate.
// 3. Load all the Events from snap.version and process them on Aggregate
// 4. Take snapshot of Aggregate
// 5. Save snapshot of Aggregate with version = Aggregate.version

// s.factory
func (s *snapshotter) process() {
	//log.Info("cqrs.snapshot", "running...")

	//1. load aggregates... snapshooter
	aggregates, err := s.events.store.Last(s.kind, s.frequency)
	if err != nil {
		log.Error("cqrs.snapshot.load-last", err)
		return
	}

	for _, a := range aggregates {
		aggregate := s.factory()
		aggregate.Root().init(a.ID, a.Version)

		//2. load last snap and restore aggregate with given snapshot data
		version, data := s.events.store.Snapshot(aggregate.Root().ID)
		if len(data) > 0 {
			snapshot, err := s.serializer.Unmarshal(s.snapshot.Name, data)
			if err != nil {
				log.Error("cqrs.snapshot.unmarshal", err)
				continue
			}

			if err := aggregate.RestoreSnapshot(snapshot); err != nil {
				log.Error("cqrs.snapshot.restore", err)
				continue
			}
		}

		// 3. Load all the Events from snap.version and process them on Aggregate
		num, err := s.events.load(aggregate, version)
		if err != nil {
			log.Error("cqrs.snapshot.events-load", err)
			continue
		}

		// 4. Take snapshot of Aggregate
		data, err = s.serializer.Marshal(s.snapshot.Name, aggregate.TakeSnapshot())
		if err != nil {
			log.Error("cqrs.snapshot.marshal", err)
			continue
		}

		// 5. Save snapshot of Aggregate with version = Aggregate.version
		v := Snapshot{aggregate.Root().ID, data, aggregate.Root().Version}
		if err = s.events.store.Make(v); err != nil {
			log.Error("cqrs.snapshot.make", err)
			continue
		}

		log.Info(
			"cqrs.snapshot", "%s.#%s.v%d taken, rebuilded "+
				"from .v%d, with %d processed events",
			a.Type, a.ID[24:], a.Version, version, num)
	}
}

func (s *snapshotter) Load(a Aggregate) error {

	version, data := s.events.store.Snapshot(a.Root().ID)

	// we have snapshot, restore it!
	if len(data) > 0 {
		snapshot, err := s.serializer.Unmarshal(s.snapshot.Name, data)
		if err != nil {
			return err
		}

		if err := a.RestoreSnapshot(snapshot); err != nil {
			return err
		}
	}

	events, err := s.events.load(a, version)
	if err != nil {
		return err
	}

	log.Info("cqrs.snapshot",
		"%s.#%s.v%d restored from .v%d snapshot with %d events",
		a.Root().Type, a.Root().ID[24:], a.Root().Version, version, events)

	return nil
}

func newSnapshotter(frequency uint, e *events, f func() Aggregate) *snapshotter {
	a := f()
	sStruct := a.TakeSnapshot()
	return &snapshotter{
		frequency:  frequency,
		events:     e,
		factory:    f,
		kind:       a.Root().Type,
		snapshot:   newStructure(sStruct),
		serializer: newSerializer(sStruct),
	}
}
