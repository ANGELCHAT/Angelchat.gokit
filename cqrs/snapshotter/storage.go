package snapshotter

type Storage interface {
	Load(string) (Snapshot, error)
	Replace(Snapshot) error
}

type mem struct {
	list map[string]Snapshot
}

func (m *mem) Load(id string) (Snapshot, error) {
	s, ok := m.list[id]
	if !ok {
		return Snapshot{}, nil
	}

	return s, nil
}

func (m *mem) Replace(s Snapshot) error {
	m.list[s.ID] = s
	return nil
}

func NewMemoryStorage() *mem {
	return &mem{
		list: make(map[string]Snapshot),
	}
}
