package example

import (
	"fmt"
	"time"

	"reflect"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/example/events"
)

type subscription struct {
	person string
	meal   string
	on     time.Time
}

type aggregate struct {
	//cqrs aggregate
	root *cqrs.Aggregate

	//Business data
	name          string
	info          string
	menu          []string
	subscriptions map[string]subscription
	scheduled     time.Time
	canceled      time.Time
}

//
// Business Methods
//
func (a *aggregate) Subscribe(person, meal string) error {
	if !a.canceled.IsZero() {
		return fmt.Errorf("%s subscriptions has been canceled", a.name)
	}

	d := time.Now()
	s, ok := a.subscriptions[person]

	if ok {
		a.root.Apply(&events.MealChanged{
			Person:      person,
			PreviewMeal: s.meal,
			ActualMeal:  meal,
			At:          d})

		return nil
	}

	a.root.Apply(&events.MealSelected{
		Person: person,
		Meal:   meal,
		At:     d})

	return nil
}

func (a *aggregate) Reschedule(date time.Time) error {
	if !a.canceled.IsZero() {
		return fmt.Errorf("%s is canceled", a.name)
	}

	a.root.Apply(&events.Rescheduled{On: date})

	return nil
}

func (a *aggregate) Schedule(date time.Time) error {
	if !date.After(time.Now()) {
		return fmt.Errorf("tavern %s can not be scheduled in past", a.name)
	}

	if !a.canceled.IsZero() {
		return fmt.Errorf("tavern %s has been canceled", a.name)
	}

	if !a.scheduled.IsZero() {
		return fmt.Errorf(
			"tavern %s is already scheduled for %s",
			a.name, a.scheduled.Format("2006-01-02"))
	}

	a.root.Apply(&events.Scheduled{On: date})

	return nil
}

func (a *aggregate) Cancel() error {
	var people []string

	if !a.canceled.IsZero() {
		return fmt.Errorf("%s already canceled", a.name)
	}

	for _, p := range a.subscriptions {
		people = append(people, p.person)
	}

	a.root.Apply(&events.Canceled{
		Restaurant: a.name,
		People:     people,
		At:         time.Now()})

	return nil
}

func (a *aggregate) Create(name, info string, menu ...string) error {
	a.root.Apply(&events.Created{
		Restaurant: name,
		Info:       info,
		Menu:       menu,
	})

	return nil
}

func handler(a *aggregate) func(interface{}) error {
	return func(e interface{}) error {
		switch e := e.(type) {
		case *events.Created:
			a.name, a.info, a.menu = e.Restaurant, e.Info, e.Menu
			a.subscriptions = map[string]subscription{}

		case *events.Scheduled:
			a.scheduled = e.On

		case *events.MealSelected:
			a.subscriptions[e.Person] = subscription{
				person: e.Person,
				meal:   e.Meal,
				on:     e.At,
			}

		case *events.MealChanged:
			a.subscriptions[e.Person] = subscription{
				person: e.Person,
				meal:   e.ActualMeal,
				on:     e.At,
			}

		case *events.Canceled:
			a.canceled = e.At

		default:
			return fmt.Errorf("event %s not handled", reflect.TypeOf(e))
		}

		return nil
	}
}
