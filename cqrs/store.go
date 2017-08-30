package cqrs

type Store interface {
	Save(string, []Record) error
	Load(string) ([]Record, error)
}
type memStorage struct {
	events map[string][]Record
}

func (s *memStorage) Save(id string, es []Record) error {
	s.events[id] = append(s.events[id], es...)
	return nil
}

func (s *memStorage) Load(id string) ([]Record, error) {
	return s.events[id], nil
}

func newMemStorage() Store {
	return &memStorage{
		events: map[string][]Record{},
	}
}
