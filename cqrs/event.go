package cqrs

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/sokool/gokit/log"
)

type Command interface{}

type Event2 interface{}

type Aggregate struct {
	ID       string
	Name     string
	Version  uint
	Commands map[Command]CommandHandler
	Events   map[Event2]EventHandler

	RestoreSnapshot func(v interface{}) error
	TakeSnapshot    func() interface{}

	events []Event2
}

func (a *Aggregate) String() string {
	return fmt.Sprintf("%s.#%s.v%d", a.Name, a.ID[24:], a.Version)
}

func (a *Aggregate) dispatch(c Command) error {
	name := reflect.TypeOf(c).String()
	for v, handler := range a.Commands {
		if name == reflect.TypeOf(v).String() {
			events, err := handler(c)
			if err != nil {
				return err
			}

			a.events = append(a.events, events...)

			return nil
		}
	}

	return fmt.Errorf("command handler for %s not exists", name)
}

func (a *Aggregate) apply(e Event2) {
	name := reflect.TypeOf(e).String()
	for v, handler := range a.Events {
		if handler == nil {
			continue
		}

		if name == reflect.TypeOf(v).String() {

			if err := handler(e); err != nil {
				log.Info("cqrs.event.handle", err.Error())
			} else {
				log.Info("cqrs.event.handled", "%s:%+v", name, e)
			}

			return
		}
	}

	log.Info("cqrs.event.dispatch",
		"event handler for %s not registered in cqrs.Aggregate",
		name)

}

type FactoryFunc func(string, uint) *Aggregate

type CommandHandler func(Command) ([]Event2, error)

type EventHandler func(Event2) error

//
//
//

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

//todo aggregate{ID, Name}?
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
