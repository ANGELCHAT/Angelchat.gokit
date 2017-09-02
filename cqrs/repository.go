package cqrs

import (
	"time"

	"github.com/sokool/gokit/log"
)

type Repository struct {
	serializer *serializer
	opts       *Options
	aggregate  structure
	handler    func(interface{}) error
}

func (s *Repository) Save(root *Root) error {
	var events []Event
	var id Identity = root.ID

	version := root.Version
	for _, o := range root.events {
		structure := newStructure(o)
		data, err := s.serializer.Marshal(structure.Name, o)
		if err != nil {
			return err
		}

		version++
		events = append(events, Event{
			ID:      generateIdentity(),
			Type:    structure.Name,
			Data:    data,
			Created: time.Now(),
			Version: version,
		})
	}

	if len(id) == 0 {
		id = generateIdentity()
	}

	aggregate := Aggregate{
		ID:      id.String(),
		Type:    root.Type,
		Version: version,
	}

	// store aggregate state
	if err := s.opts.Storage.Save(aggregate, events); err != nil {
		return err
	}

	log.Debug("cqrs.save.aggregate", "%s", aggregate.String())

	// send events to listeners of aggregate
	if s.opts.Handlers != nil {
		for _, eh := range s.opts.Handlers {
			eh(aggregate, events, root.events)
		}
	}

	root.ID = id
	root.events = []interface{}{}
	root.Version = version

	return nil
}

func (s *Repository) Load(id string, h func(interface{}) error) (*Root, error) {
	agg, events, err := s.opts.Storage.Load(id)
	if err != nil {
		return nil, err
	}
	log.Debug("cqrs.load.aggregate", "%s", agg.String())

	root := &Root{
		ID:      Identity(agg.ID),
		Version: agg.Version,
		Type:    agg.Type,
		events:  []interface{}{},
		handler: h,
	}

	for _, event := range events {
		log.Debug("cqrs.load.event", event.String())
		e, err := s.serializer.Unmarshal(event.Type, event.Data)
		if err != nil {
			log.Error("cqrs.load.event", err)
			return root, err
		}

		if err := root.handle(e); err != nil {
			log.Error("cqrs.handle.event", err)
			return root, err
		}
	}

	return root, nil
}

func (s *Repository) Aggregate() interface{} {
	return s.aggregate.Instance()
}

//todo - required "aggregate name", "list of events"
func New(es []interface{}, os ...Option) *Repository {
	return &Repository{
		serializer: newSerializer(es...),
		opts:       newOptions(os...),
		handler:    func(interface{}) error { return nil },
	}
}
