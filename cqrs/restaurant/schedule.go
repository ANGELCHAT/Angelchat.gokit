package restaurant

import (
	"fmt"
	"reflect"
	"time"

	"github.com/sokool/gokit/cqrs"
)

type Schedule struct{ On time.Time }
type Reschedule struct{ On time.Time }
type EventScheduled struct{ On time.Time }
type EventRescheduled struct{ On time.Time }

func reschedule(r *Restaurant) cqrs.CommandHandler {
	return func(v cqrs.Command) ([]cqrs.Event2, error) {
		c, ok := v.(*Reschedule)
		if !ok {
			return nil, fmt.Errorf("wrong %s command type", reflect.TypeOf(v))
		}

		if !r.Canceled.IsZero() {
			return nil, fmt.Errorf("%s is canceled", r.Name)
		}

		return []cqrs.Event2{&EventRescheduled{On: c.On}}, nil
	}
}

func schedule(r *Restaurant) cqrs.CommandHandler {
	return func(v cqrs.Command) ([]cqrs.Event2, error) {
		c, ok := v.(*Schedule)
		if !ok {
			return nil, fmt.Errorf("wrong %s command type", reflect.TypeOf(v))
		}

		if !c.On.After(time.Now()) {
			return nil, fmt.Errorf("restaurant %s can not be scheduled in past", r.Name)
		}

		if !r.Canceled.IsZero() {
			return nil, fmt.Errorf("restaurant %s has been canceled", r.Name)
		}

		if !r.Scheduled.IsZero() {
			return nil, fmt.Errorf(
				"restaurant %s is already scheduled for %s",
				r.Name, r.Scheduled.Format("2006-01-02"))
		}

		return []cqrs.Event2{&EventScheduled{On: c.On}}, nil
	}
}

func scheduled(r *Restaurant) cqrs.EventHandler {
	return func(v cqrs.Event2) error {
		switch e := v.(type) {
		case *EventScheduled:
			r.Scheduled = e.On

		case *EventRescheduled:
			r.Scheduled = e.On
		default:
			return fmt.Errorf("wrong %s command type", reflect.TypeOf(v))
		}

		return nil
	}
}
