package cqrs

import "fmt"

type Service struct {
	serializer *serializer
	opts       *Options
}

func (s *Service) Save(id string, events ...Event) error {
	var records []Record
	for _, e := range events {
		o, err := s.serializer.Marshal(e)
		if err != nil {
			return err
		}

		records = append(records, o)
	}

	if err := s.opts.Storage.Save(id, records); err != nil {
		return err
	}

	if s.opts.Handlers != nil {
		go func(es []Event) {
			for _, e := range es {
				for _, h := range s.opts.Handlers {
					h(e)
				}
			}
		}(events)
	}

	return nil
}

func (s *Service) Load(id string, h HandlerFunc) error {
	rs, err := s.opts.Storage.Load(id)
	if err != nil {
		return err
	}

	if len(rs) == 0 {
		return fmt.Errorf("#%s %s not found", id, s.opts.Name)
	}

	for _, record := range rs {
		event, err := s.serializer.Unmarshal(record)
		if err != nil {
			return err
		}

		if err := h(event); err != nil {
			return err
		}
	}

	return nil
}

func New(es []Event, os ...Option) *Service {
	return &Service{
		serializer: newSerializer(es...),
		opts:       newOptions(os...),
	}
}
