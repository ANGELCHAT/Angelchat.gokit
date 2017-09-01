package cqrs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/sokool/gokit/log"
)

type Identity string

func (i Identity) String() string {
	return string(i)
}

func generateIdentity() Identity {
	return Identity(uuid.New().String())
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
	ID      Identity
	Type    string
	Data    []byte
	Version uint64
	Created time.Time
}

func (r Event) String() string {
	return fmt.Sprintf("%s #%d.%s%s",
		r.ID[24:], r.Version, r.Type, r.Data)
}

type Root struct {
	ID      Identity
	Name    string
	version uint64
	events  []interface{}
	handler func(interface{}) error
}

func (a *Root) Identity() Identity {
	if len(a.ID) == 0 {
		a.ID = generateIdentity()
	}
	return a.ID
}

func (a *Root) Apply(e interface{}) error {
	if err := a.handle(e); err != nil {
		log.Error("tavern.event.handling", err)
		return err
	}

	a.events = append(a.events, e)
	return nil
}

func (a *Root) handle(v interface{}) error {
	return a.handler(v)
}

func NewAggregate(name string, f func(interface{}) error) *Root {
	return &Root{
		Name:    name,
		events:  []interface{}{},
		handler: f,
	}
}

//todo maybe interface?
type Aggregate struct {
	ID      string
	Name    string
	Version uint64
	Events  []Event
}
