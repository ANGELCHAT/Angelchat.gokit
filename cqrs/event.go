package cqrs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/sokool/gokit/log"
)

type Aggregate interface {
	Root() *Root
	Set(*Root)

	// todo separate interface Snapshooter? consider as event?
	TakeSnapshot() interface{}
	RestoreSnapshot(interface{}) error
}

type Factory func() (Aggregate, DataHandler)

type DataHandler func(e interface{}) error

func generateID() string {
	return uuid.New().String()
}

type structure struct {
	Name string
	Type reflect.Type
}

func (i structure) Instance() interface{} {
	return reflect.New(i.Type).Interface()
}

func newStructure(v interface{}) structure {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return structure{t.Name(), t}
}

type Event struct {
	ID      string
	Type    string
	Data    []byte
	Version uint64
	Created time.Time
}

func (e Event) String() string {
	return fmt.Sprintf("#%s: v%d.%s%s",
		e.ID[24:], e.Version, e.Type, e.Data)
}

//todo Root == CQRSAggregate???
type Root struct {
	ID      string
	Version uint64
	Type    string
	events  []interface{}
	handler func(interface{}) error
}

func (a *Root) init(id string, version uint64) {
	a.ID = id
	a.Version = version
	a.events = []interface{}{}
}

func (a *Root) Apply(e interface{}) error {
	if err := a.handler(e); err != nil {
		log.Error("tavern.event.handling", err)
		return err
	}
	a.events = append(a.events, e)
	return nil
}

func (a *Root) String() string {
	return fmt.Sprintf("#%s: v%d.%s:ROOT", a.ID[24:], a.Version, a.Type)
}
func newRoot(h DataHandler, name string) *Root {
	return &Root{
		Type:    name,
		events:  []interface{}{},
		handler: h,
	}
}

//todo maybe interface?
type CQRSAggregate struct {
	ID      string
	Type    string
	Version uint64
}

func (a *CQRSAggregate) String() string {
	return fmt.Sprintf("#%s: v%d.%s:AGGREGATE",
		a.ID[24:], a.Version, a.Type)
}

type Snapshot struct {
	AggregateID string
	Data        []byte
	Version     uint64
}
