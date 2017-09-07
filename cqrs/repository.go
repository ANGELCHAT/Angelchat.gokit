package cqrs

import (
	"time"

	"github.com/sokool/gokit/log"
)

type Repository struct {
	name        string
	serializer  *serializer
	factory     Factory
	opts        *Options
	snapshotter *Snapshotter
}

func (s *Repository) Aggregate() Aggregate {
	return s.aggregateInstance("", 0)
}

func (s *Repository) Save(a Aggregate) error {
	var r *Root = a.Root()
	var events []Event
	var aggregate = CQRSAggregate{
		ID:      r.ID,
		Type:    s.name,
		Version: r.Version,
	}

	if len(aggregate.ID) == 0 {
		aggregate.ID = generateID()
	}

	log.Info("cqrs.save.aggregate", "%s with %d new events",
		aggregate.String(), len(r.events))

	for _, o := range r.events {
		structure := newStructure(o)
		data, err := s.serializer.Marshal(structure.Name, o)
		if err != nil {
			log.Error("cqrs.save.event", err)
			return err
		}

		aggregate.Version++
		events = append(events, Event{
			ID:      generateID(),
			Type:    structure.Name,
			Data:    data,
			Created: time.Now(),
			Version: aggregate.Version,
		})

		//log.Debug("cqrs.save.event", events[i].String())
	}

	// store aggregate state
	if err := s.opts.Storage.Save(aggregate, events); err != nil {
		log.Error("cqrs.save.aggregate", err)
		return err
	}

	// send events to listeners of aggregate
	if s.opts.Handlers != nil {
		for _, eh := range s.opts.Handlers {
			eh(aggregate, events, r.events)
		}
	}

	r.init(aggregate.ID, aggregate.Version)
	r.events = []interface{}{}

	return nil
}

func (s *Repository) Load(id string) (Aggregate, error) {
	var version uint64
	var aggregate Aggregate
	var err error

	// take events from last snapshot?
	if s.snapshotter != nil {
		aggregate, err = s.snapshotter.Load(id)
		if err != nil {
			return nil, err
		}
	} else {
		a, err := s.opts.Storage.Load(id)
		if err != nil {
			return nil, err
		}

		aggregate = s.aggregateInstance(a.ID, a.Version)
	}

	//// check if aggregate is stored.
	//ca, err := s.opts.Storage.Load(id)
	//if err != nil {
	//	return nil, err
	//}
	//
	log.Info("cqrs.load.aggregate", "%s", aggregate.Root().String())

	// load aggregate events from given version.

	// todo do not load from last snapshotted version until user decide to!
	events, err := s.opts.Storage.Events(version, id)
	if err != nil {
		return nil, err
	}
	var event Event
	for _, event = range events {
		e, err := s.serializer.Unmarshal(event.Type, event.Data)
		if err != nil {
			log.Error("cqrs.load.event", err)
			return nil, err
		}

		if err := aggregate.Root().handler(e); err != nil {
			log.Error("cqrs.handle.event", err)
			return nil, err
		}
	}

	return aggregate, nil
}

func (s *Repository) aggregateInstance(id string, version uint64) Aggregate {
	a, h := s.factory()
	r := newRoot(h, s.name)
	r.init(id, version)
	a.Set(r)

	return a
}

//todo return error
func (s *Repository) Snapshotter(everyVersion uint, frequency time.Duration) {
	if s.snapshotter != nil {

		return
	}

	s.snapshotter = NewSnapshotter(s.name, everyVersion, s.opts.Storage, s)
	timer := time.NewTicker(frequency)

	go func(t *time.Ticker) {
		log.Info("cqrs.snapshot.start", "%s, every %s and %d version",
			s.name, frequency, everyVersion)

		//todo break that loop
		for range t.C {
			s.snapshotter.Run()
		}

		log.Info("cqrs.snapshot.stop", s.name)
	}(timer)

}

func NewRepository(f Factory, es []interface{}, os ...Option) *Repository {
	aggregate, _ := f()

	return &Repository{
		serializer: newSerializer(es...),
		opts:       newOptions(os...),
		factory:    f,
		name:       newStructure(aggregate).Name,
	}
}
