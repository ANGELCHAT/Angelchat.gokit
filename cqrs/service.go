package cqrs

import (
	"fmt"
	"time"
)

type Service struct {
	serializer *serializer
	opts       *Options
}

func (s *Service) Save(a *Aggregate) error {
	var records []Event
	events := map[*Event]interface{}{}

	for _, o := range a.events {
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
			Version: 0,
		}

		events[&e] = o
		records = append(records, e)
	}

	if err := s.opts.Storage.Save(a.Identity(), records); err != nil {
		return err
	}

	if s.opts.Handlers != nil {
		for r, o := range events {
			for _, h := range s.opts.Handlers {
				h(a.ID, *r, o)
			}
		}
	}

	return nil
}

func (s *Service) Load(a *Aggregate) error {
	rs, err := s.opts.Storage.Load(a.Identity())
	if err != nil {
		return err
	}

	if len(rs) == 0 {
		return fmt.Errorf("#%s %s not found", a.Identity(), a.Name)
	}

	for _, event := range rs {
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
