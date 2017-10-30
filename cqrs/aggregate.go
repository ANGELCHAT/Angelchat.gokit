package cqrs

import (
	"fmt"
	"reflect"

	"github.com/sokool/gokit/log"
)

type FactoryFunc func(string, uint) *Aggregate

type Aggregate struct {
	ID       string
	Name     string
	Version  uint
	Commands map[Command]CommandHandler
	Events   map[interface{}]EventHandler

	RestoreSnapshot func(v interface{}) error
	TakeSnapshot    func() interface{}

	events []interface{}
}

func (a *Aggregate) String() string {
	return fmt.Sprintf("%s.#%s.v%d", a.Name, a.ID[24:], a.Version)
}

func (a *Aggregate) dispatch(c Command) error {
	name := reflect.TypeOf(c).String()
	log.Debug("cqrs.aggregate.dispatch", "%s command", name)
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

func (a *Aggregate) apply() {
	log.Debug("cqrs.aggregate.apply", "%d events received", len(a.events))
	for _, e := range a.events {
		en := reflect.TypeOf(e).String()
		for v, handler := range a.Events {
			if handler == nil {
				continue
			}

			vn := reflect.TypeOf(v).String()
			if en == vn {
				if err := handler(e); err != nil {
					log.Error("cqrs.aggregate.apply", err)
					break
				}

				//log.Debug("cqrs.aggregate.apply", "%s handled", en)
				break
			}
		}
	}
	//a.events = make([]interface{}, 0)

}
