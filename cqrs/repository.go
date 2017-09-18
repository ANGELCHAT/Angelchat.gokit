package cqrs

import "github.com/sokool/gokit/log"

type Repository struct {
	name        string
	factory     Factory
	opts        *Options
	snapshotter *snapshotter

	cache  *cache
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

	if s.cache != nil {
		s.cache.store(a)
	}

	// send events to listeners of aggregate
	//if s.opts.Handlers != nil {
	//	for _, eh := range s.opts.Handlers {
	//		eh(aggregate, events, r.events)
	//	}
	//}

	return nil
}

func (s *Repository) Load(id string) (Aggregate, error) {

	// load aggregate from cache
	if s.cache != nil {
		aggregate, has := s.cache.restore(id)
		if has {
			return aggregate, nil
		}
	}

	stored, err := s.opts.Store.Load(id)
	if err != nil {
		return nil, err
	}

	var aggregate = s.aggregateInstance()
	aggregate.Root().init(stored.ID, stored.Version)

	// then restore aggregate from snapshotter
	if s.snapshotter != nil {
		if err = s.snapshotter.restore(aggregate); err != nil {
			return nil, err
		}

		return aggregate, nil
	}

	// load aggregate events from given version.
	events, err := s.events.load(aggregate, 0)
	if err != nil {
		return aggregate, err
	}

	log.Info("cqrs.repository",
		"%s reconstructed from beginning by processing %d events",
		aggregate.Root(), events)

	return aggregate, nil
}

func (s *Repository) aggregateInstance() Aggregate {
	aggregate, handler := s.factory()
	aggregate.Set(newRoot(handler, s.name))

	return aggregate
}

func NewRepository(f Factory, es []interface{}, os ...Option) *Repository {
	aggregate, _ := f()
	options := newOptions(os...)
	r := &Repository{
		opts:    options,
		factory: f,
		name:    newStructure(aggregate).Name,
		events: &events{
			store:      options.Store,
			serializer: newSerializer(es...),
		},
	}

	if options.SnapEpoch > 0 && options.SnapFrequency > 0 {
		r.snapshotter = newSnapshotter(options.SnapEpoch, r.events, r.aggregateInstance)
		go r.snapshotter.run(options.SnapFrequency)
	}

	if options.Cache {
		r.cache = newCache()
	}

	return r
}
