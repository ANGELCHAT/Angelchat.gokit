package cqrs

import (
	"time"

	"github.com/sokool/gokit/log"
)

type Repository struct {
	name        string
	factory     Factory
	opts        *Options
	snapshotter *snapshotter

	events *events
}

func (s *Repository) Aggregate() Aggregate {
	aggregate := s.aggregateInstance()
	aggregate.Root().init("", 0)
	return aggregate
}

func (s *Repository) Save(a Aggregate) error {

	events, err := s.events.save(a)
	if err != nil {
		log.Error("cqrs.repository", err)
		return err
	}

	log.Info("cqrs.repository", "%s saved with %d new events",
		a.Root().String(), events)

	// send events to listeners of aggregate
	//if s.opts.Handlers != nil {
	//	for _, eh := range s.opts.Handlers {
	//		eh(aggregate, events, r.events)
	//	}
	//}

	return nil
}

func (s *Repository) Load(id string) (Aggregate, error) {
	var aggregate = s.aggregateInstance()
	var err error

	a, err := s.opts.Storage.Load(id)
	if err != nil {
		return nil, err
	}

	aggregate.Root().init(a.ID, a.Version)

	// take aggregate from last snapshot?
	if s.snapshotter != nil {
		if err = s.snapshotter.Load(aggregate); err != nil {
			return nil, err
		}

		return aggregate, nil
	}

	// load aggregate events from given version.
	events, err := s.events.load(aggregate, 0)
	if err != nil {
		return aggregate, err
	}

	log.Info("cqrs.repository", "%s loaded, rebuilded from beginning by processing %d events",
		aggregate.Root(), events)

	return aggregate, nil
}

func (s *Repository) aggregateInstance() Aggregate {
	aggregate, handler := s.factory()
	aggregate.Set(newRoot(handler, s.name))

	return aggregate
}

//todo return error
func (s *Repository) Snapshotter(everyVersion uint, frequency time.Duration) {
	if s.snapshotter != nil {
		return
	}

	s.snapshotter = newSnapshotter(everyVersion, s.events, s.aggregateInstance)
	timer := time.NewTicker(frequency)

	go func(t *time.Ticker) {
		log.Info("cqrs.snapshot", "starting %s snapshotter every %s and every %d version",
			s.name, frequency, everyVersion)

		//todo break that loop
		for range t.C {
			s.snapshotter.process()
		}

		log.Info("cqrs.snapshot", s.name)
	}(timer)

}

func NewRepository(f Factory, es []interface{}, os ...Option) *Repository {
	aggregate, _ := f()
	options := newOptions(os...)

	return &Repository{
		opts:    options,
		factory: f,
		name:    newStructure(aggregate).Name,
		events: &events{
			store:      options.Storage,
			serializer: newSerializer(es...),
		},
	}
}
