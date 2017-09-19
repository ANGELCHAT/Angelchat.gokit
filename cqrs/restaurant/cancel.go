package restaurant

import (
	"fmt"
	"time"

	"reflect"

	"github.com/sokool/gokit/cqrs"
)

type EventCanceled struct {
	Restaurant string
	People     []string
	At         time.Time
}

type Cancel struct{}

func cancel(r *Restaurant) cqrs.CommandHandler {
	return func(v cqrs.Command) ([]cqrs.Event2, error) {
		_, ok := v.(*Cancel)
		if !ok {
			return nil, fmt.Errorf("wrong %s command type", reflect.TypeOf(v))
		}

		var people []string

		if r.Created.IsZero() {
			return nil, fmt.Errorf("not created yet")
		}

		if !r.Canceled.IsZero() {
			return nil, fmt.Errorf("%s already canceled", r.Name)
		}

		for _, p := range r.Subscriptions {
			people = append(people, p.Person)
		}

		events := []cqrs.Event2{
			&EventCanceled{
				Restaurant: r.Name,
				People:     people,
				At:         time.Now()}}

		return events, nil
	}
}

func canceled(r *Restaurant) cqrs.EventHandler {
	return func(v cqrs.Event2) error {
		e := v.(*EventCanceled)

		r.Canceled = e.At

		return nil
	}
}
