package es

//import (
//	"fmt"
//	"reflect"
//	"time"
//
//	"github.com/google/uuid"
//	"github.com/sokool/gokit/cqrs/platform"
//	"github.com/sokool/gokit/locker"
//	"github.com/sokool/gokit/log"
//)
//
//type EventStore struct {
//	access     *locker.Key
//	serializer *platform.Serializer
//	opts       *Options
//}
//
//var (
//	ErrNoAggregateIdentity = fmt.Errorf("aggregate has no identity")
//	ErrWrongAggregateName  = fmt.Errorf("wrong aggregate name")
//	ErrEmptyEvents  = fmt.Errorf("empty events")
//)
//
//func (s *EventStore) Load(aggregate string, from uint) ([]interface{}, uint, error) {
//	s.access.Lock(aggregate)
//	defer s.access.Unlock(aggregate)
//	if from <=1 {
//		from = 1
//	}
//
//	version := from
//	stream := make([]interface{}, 0)
//	events, err := s.opts.storage.Load(aggregate, from)
//	if err != nil {
//		return stream, version, err
//	}
//
//	for i, event := range events {
//		if event.Version != from+uint(i) {
//			return stream, version, fmt.Errorf(
//				"wrong %s events order detected, has %d, expected %d",
//				event.AggregateType, event.Version, from+uint(i))
//		}
//
//		log.Debug("es.repository.load", "deserialize %s v.%d", event.Type, event.Version)
//		e, err := s.serializer.Unmarshal(event.Type, event.Data)
//		if err != nil {
//			return stream, 0, err
//		}
//
//		stream = append(stream, e)
//		version = event.Version
//	}
//
//	return stream, version, nil
//}
//
//func (s *EventStore) Save(aggregate, name string, es ...interface{}) (uint, error) {
//	if len(aggregate) == 0 {
//		return 0, ErrNoAggregateIdentity
//	}
//
//	if len(name) == 0 {
//		return 0, ErrWrongAggregateName
//	}
//
//	if len(es) == 0 {
//		return 0, ErrEmptyEvents
//	}
//
//	s.access.Lock(aggregate)
//	defer s.access.Unlock(aggregate)
//
//	var events []Event
//	var date = time.Now()
//	var version uint
//
//	// determine version of aggregate
//	re, ok := s.opts.storage.Recent(aggregate)
//	if ok {
//		version = re.Version
//	}
//
//	// prepare events for storage
//	for _, event := range es {
//		version++
//
//		t := reflect.TypeOf(event)
//		if t.Kind() == reflect.Ptr {
//			t = t.Elem()
//		}
//
//		bytes, err := s.serializer.Marshal(t.Name(), event)
//		if err != nil {
//			return version, err
//		}
//
//		events = append(events, Event{
//			ID:            uuid.New().String(),
//			Type:          t.Name(),
//			AggregateID:   aggregate,
//			AggregateType: name,
//			Data:          bytes,
//			Created:       date, // each event in transaction get the same date
//			Version:       version,
//		})
//
//		log.Debug("es.repository.save", "marshaled %s v.%d", t.Name(), version)
//	}
//
//	// store aggregate state and it's events
//	if err := s.opts.storage.Append(aggregate, events); err != nil {
//		return version, err
//	}
//
//	if len(s.opts.Listener) > 0 {
//		for _, e := range events {
//			for _, l := range s.opts.Listener {
//				l(e)
//			}
//		}
//	}
//
//	return version, nil
//}
//
//func NewRepository(events []interface{}, os ...Option) *EventStore {
//	return &EventStore{
//		serializer: platform.NewSerializer(events...),
//		access:     locker.NewKey(),
//		opts:       newOptions(os...),
//	}
//}
