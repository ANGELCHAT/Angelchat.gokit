package store

//type EventStore struct {
//	Persistence Persistence
//	serializer  *serializer
//}
//
//func (s *EventStore) Snapshot(aggregate string) {
//	//v, data := s.Persistence.Snapshot(aggregate)
//	//s.serializer.Unmarshal(, data)
//
//}
//
//func (s *EventStore) Aggregate(id string) (Aggregate, error) {
//	return s.Persistence.Load(id)
//}

//func (s *EventStore) Load(a Aggregate) ([]interface{}, error) {
//	stream := make([]interface{}, 0)
//	events, err := s.Persistence.Events(a.ID, a.Version)
//	if err != nil {
//		return stream, err
//	}
//
//	for _, event := range events {
//		e, err := s.serializer.Unmarshal(event.Type, event.Data)
//		if err != nil {
//			return stream, err
//		}
//
//		stream = append(stream, e)
//	}
//
//	return stream, nil
//}

// Takes events from Aggregate.
// Store aggregate and new events.
// Increase Aggregate version +len(root.events).
// Clear Aggregate events buffer
//func (s *EventStore) SaveEvents(a *Aggregate, es ...interface{}) ([]Event, error) {
//	var events []Event
//	var date = time.Now()
//	//var version = a.Version
//	//var id = a.ID
//
//	// check if stored Aggregate has not been changed.
//	if err := s.changed(*a); err != nil {
//		return events, err
//	}
//
//	// prepare events for storage
//	for _, event := range es {
//		a.Version++
//
//		t := reflect.TypeOf(event)
//		if t.Kind() == reflect.Ptr {
//			t = t.Elem()
//		}
//
//		data, err := s.serializer.Marshal(t.Name(), event)
//		if err != nil {
//			return events, err
//		}
//
//		events = append(events, Event{
//			ID:      uuid.New().String(),
//			Type:    t.Name(),
//			Data:    data,
//			Created: date, // each event in transaction get the same date
//			Version: a.Version,
//		})
//
//	}
//
//	//todo move this out from save scope
//	// generate Aggregate ID if empty
//	//if len(a.ID) == 0 {
//	//	id = generateID()
//	//}
//
//	//log.Info("events.store", "", root.Type)
//	// store aggregate state and it's events
//	//v := Aggregate{id, a.Name, version}
//	if err := s.Persistence.Save(*a, events); err != nil {
//		return events, err
//	}
//
//	// modify Aggregate state
//	//a.ID = id
//	//a.Version = version
//	//a.events = []Event2{} //todo do not clear that at this tage
//
//	return events, nil
//}

//func (s *EventStore) changed(a Aggregate) error {
//
//	if len(a.ID) == 0 {
//		return nil
//	}
//
//	stored, err := s.Persistence.Load(a.ID)
//	if err != nil {
//		return err
//	}
//
//	if uint(stored.Version) != a.Version {
//		return fmt.Errorf("%s.#%s transaction failed, current version is %d, but stored is %d",
//			a.Type, a.ID[24:], a.Version, stored.Version)
//	}
//
//	return nil
//}

//func NewStorage(es []interface{}, snap interface{}) *EventStore {
//	var v []interface{}
//	copy(v, es)
//	v = append(v, snap)
//
//	return &EventStore{
//		Persistence: NewMemoryStorage(),
//		serializer:  newSerializer(v),
//	}
//}
