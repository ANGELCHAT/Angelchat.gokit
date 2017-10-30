package cqrs

import (
	"fmt"
	"time"
)

type Command interface{}

type Event2 interface{}

type CommandHandler func(Command) ([]interface{}, error)

type EventHandler func(Event2) error

//
//
//

//func generateID() string {
//	return uuid.New().String()
//}
//
//type structure struct {
//	Name string
//	Type reflect.Type
//}
//
//func (i structure) Instance() interface{} {
//	return reflect.New(i.Type).Interface()
//}
//
//func newStructure(v interface{}) structure {
//	t := reflect.TypeOf(v)
//	if t.Kind() == reflect.Ptr {
//		t = t.Elem()
//	}
//
//	return structure{t.Name(), t}
//}

//todo aggregate{ID, Name}?
//todo it belongs to EVENT STORE!
type Event struct {
	ID      string
	Type    string
	Data    []byte
	Version uint
	Created time.Time
}

func (e Event) String() string {
	return fmt.Sprintf("#%s: v%d.%s%s",
		e.ID[24:], e.Version, e.Type, e.Data)
}

//todo maybe interface?
type CQRSAggregate struct {
	ID      string
	Type    string
	Version uint
}

func (a *CQRSAggregate) String() string {
	return fmt.Sprintf("#%s: v%d.%s",
		a.ID[24:], a.Version, a.Type)
}

type Snapshot struct {
	AggregateID string
	Data        []byte
	Version     uint
}
