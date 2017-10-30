package restaurant

import (
	"fmt"
	"reflect"
	"time"

	"github.com/sokool/gokit/cqrs"
)

type Create struct {
	Name string
	Info string
	Menu []string
}

type EventCreated struct {
	Restaurant string
	Info       string
	Menu       []string
	At         time.Time
}

func CreateHandler(r *Restaurant) cqrs.CommandHandler {
	return func(v cqrs.Command) ([]interface{}, error) {
		c, ok := v.(*Create)
		if !ok {
			return nil, fmt.Errorf("wrong %s command type", reflect.TypeOf(v))
		}

		if !r.Created.IsZero() {
			return nil, fmt.Errorf("restaurant %s is already created", r.Name)
		}

		es := []interface{}{
			&EventCreated{
				Restaurant: c.Name,
				Info:       c.Info,
				Menu:       c.Menu,
				At:         time.Now(),
			}}

		return es, nil
	}
}

func created(r *Restaurant) cqrs.EventHandler {
	return func(v cqrs.Event2) error {
		e, ok := v.(*EventCreated)
		if !ok {
			return fmt.Errorf("wrong event %s type", reflect.TypeOf(v))
		}

		r.Name, r.Info, r.Menu = e.Restaurant, e.Info, e.Menu
		r.Subscriptions = map[string]subscription{}
		r.Created = e.At

		return nil
	}
}
