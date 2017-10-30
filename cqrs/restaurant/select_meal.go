package restaurant

import (
	"fmt"
	"reflect"
	"time"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/log"
)

//COMMAND
type SelectMeal struct {
	Person  string
	Meal    string
	HowLong time.Duration
}

func (s SelectMeal) String() string {
	return fmt.Sprintf("%s:%s |%s", s.Person, s.Meal, s.HowLong)
}

//EVENT
type EventMealSelected struct {
	Person string
	Meal   string
	At     time.Time
}

//EVENT
type EventMealChanged struct {
	Person       string
	PreviousMeal string
	ActualMeal   string
	At           time.Time
}

func selectMeal(r *Restaurant) cqrs.CommandHandler {
	return func(v cqrs.Command) ([]interface{}, error) {
		var es []interface{}
		c, ok := v.(*SelectMeal)
		if !ok {
			return es, fmt.Errorf("wrong %s command type", reflect.TypeOf(v))
		}

		if !r.Canceled.IsZero() {
			return es, fmt.Errorf("%s subscriptions has been canceled", r.Name)
		}

		if c.HowLong > time.Duration(0) {
			log.Info("restaurant.select-meal", "%s: %s", r.ID, c.String())
			time.Sleep(c.HowLong)
		}

		s, ok := r.Subscriptions[c.Person]

		if ok {
			es = append(es, &EventMealChanged{
				Person:       c.Person,
				PreviousMeal: s.Meal,
				ActualMeal:   c.Meal,
				At:           time.Now()})

			return es, nil
		}

		es = append(es, &EventMealSelected{
			Person: c.Person,
			Meal:   c.Meal,
			At:     time.Now()})

		return es, nil
	}
}

func mealSelected(r *Restaurant) cqrs.EventHandler {
	return func(v cqrs.Event2) error {
		e, ok := v.(*EventMealSelected)
		if !ok {
			return fmt.Errorf("wrong event %s type", reflect.TypeOf(v))
		}

		r.Subscriptions[e.Person] = subscription{
			Person: e.Person,
			Meal:   e.Meal,
			On:     e.At,
		}

		return nil
	}
}

func mealChanged(r *Restaurant) cqrs.EventHandler {
	return func(v cqrs.Event2) error {
		e, ok := v.(*EventMealChanged)
		if !ok {
			return fmt.Errorf("wrong event %s type", reflect.TypeOf(v))
		}

		r.Subscriptions[e.Person] = subscription{
			Person: e.Person,
			Meal:   e.ActualMeal,
			On:     e.At,
		}

		return nil
	}
}
