package cqrs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/sokool/gokit/log"
)

//todo do not need to be exported?
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

func (e Event) String() string {
	return fmt.Sprintf("#%s: v%d.%s%s",
		e.ID[24:], e.Version, e.Type, e.Data)
}

type Root struct {
	ID      Identity
	Type    string
	Version uint64
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

func NewAggregate(name string, handler func(interface{}) error) *Root {
	return &Root{
		Type:    name,
		events:  []interface{}{},
		handler: handler,
	}
}

//todo maybe interface?
type Aggregate struct {
	ID      string
	Type    string
	Version uint64
}

func (a *Aggregate) String() string {
	return fmt.Sprintf("#%s: v%d.%s",
		a.ID[24:], a.Version, a.Type)
}
