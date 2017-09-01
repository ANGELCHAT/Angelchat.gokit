package cqrs

import "time"

type Service struct {
	serializer *serializer
	opts       *Options
}

func (s *Service) Save(a *Root) error {
	events := map[*Event]interface{}{}
	g := Aggregate{
		ID:      a.Identity().String(),
		Type:    a.Name,
		Version: a.version,
	}

	version := a.version
	for v, o := range a.events {
		version = a.version + uint64(v+1)
		structure := newStructure(o)
		data, err := s.serializer.Marshal(structure.Name, o)
		if err != nil {
			return err
		}

		e := Event{
			ID:      generateIdentity(),
			Type:    structure.Name,
			Data:    data,
			Created: time.Now(),
			Version: version,
		}
		g.Events = append(g.Events, e)
		events[&e] = o
	}

	// store aggregate state
	if err := s.opts.Storage.Save(g); err != nil {
		return err
	}

	// clear events from root
	a.events = []interface{}{}
	a.version = version

	// send events to listeners of aggregate
	if s.opts.Handlers != nil {
		for r, o := range events {
			for _, h := range s.opts.Handlers {
				h(a.ID, *r, o)
			}
		}
	}

	return nil
}

func (s *Service) Load(a *Root) error {
	agg, err := s.opts.Storage.Load(a.Identity().String())
	if err != nil {
		return err
	}

	a.ID = Identity(agg.ID)
	a.Name = agg.Type
	a.version = agg.Version

	for _, event := range agg.Events {
		e, err := s.serializer.Unmarshal(event.Type, event.Data)
		if err != nil {
			return err
		}

		if err := a.handle(e); err != nil {
			return err
		}
	}

	return nil
}

//todo - required "aggregate name", "list of events"
func New(es []interface{}, os ...Option) *Service {
	return &Service{
		serializer: newSerializer(es...),
		opts:       newOptions(os...),
	}
}
