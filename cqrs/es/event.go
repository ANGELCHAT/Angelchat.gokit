package es

import "time"

var providers = []byte("providers")
var snapshots = []byte("snapshots")
var events = []byte("events")

type Stream struct {
	ID     string
	Name   string
	Meta   []byte
	Events []Record
}

type Record struct {
	ID   string
	Name string
	Data []byte
}

type Event struct {
	ID      string
	Type    string
	Data    []byte
	Meta    []byte
	Version uint
	Created time.Time
}

type Provider struct {
	ID      string
	Type    string
	Version uint
}

type Snapshot struct {
	ID      string
	Data    []byte
	Version uint
}

type Service struct {
	store     *storage
	Publisher *Publisher
}

func (s *Service) Subscribe(su *Subscriber) error {
	return s.Publisher.Subscribe(su)
}

func (s *Service) Unsubscribe(su *Subscriber) error {
	return s.Publisher.Unsubscribe(su)
}

func (s *Service) Events(streamID string, from uint) ([]Event, error) {
	return s.store.events(streamID, from)
}

func (s *Service) Append(stream Stream) (uint, error) {
	provider, events, err := s.store.append(stream)
	if err != nil {
		return 0, err
	}

	s.Publisher.send(provider, events)

	return provider.Version, nil
}

func NewService(name string) (*Service, error) {
	s, err := newStorage(name)
	if err != nil {
		return nil, err
	}

	return &Service{
		store:     s,
		Publisher: newPublisher(),
	}, nil
}

//
//Store(Provider) error
//Provider(string) (Provider, error)
//
//SnapshottedAggregates(int) ([]Provider, error)
