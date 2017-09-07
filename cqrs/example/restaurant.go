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

type Restaurant struct {
	root *cqrs.Root

	//Business data
	name          string
	info          string
	menu          []string
	subscriptions map[string]subscription

	created   time.Time
	scheduled time.Time
	canceled  time.Time
}

// CQRS methods :/
func (a *Restaurant) Root() *cqrs.Root {
	return a.root
}

func (a *Restaurant) Set(r *cqrs.Root) {
	a.root = r
}

//
// Business Methods
//
func (a *Restaurant) Subscribe(person, meal string) error {
	if !a.canceled.IsZero() {
		return fmt.Errorf("%s subscriptions has been canceled", a.name)
	}

	d := time.Now()
	s, ok := a.subscriptions[person]

	if ok {
		a.Root().Apply(&events.MealChanged{
			Person:      person,
			PreviewMeal: s.meal,
			ActualMeal:  meal,
			At:          d})

		return nil
	}

	a.Root().Apply(&events.MealSelected{
		Person: person,
		Meal:   meal,
		At:     d})

	return nil
}

func (a *Restaurant) Reschedule(date time.Time) error {
	if !a.canceled.IsZero() {
		return fmt.Errorf("%s is canceled", a.name)
	}

	a.Root().Apply(&events.Rescheduled{On: date})

	return nil
}

func (a *Restaurant) Schedule(date time.Time) error {
	if !date.After(time.Now()) {
		return fmt.Errorf("restaurant %s can not be scheduled in past", a.name)
	}

	if !a.canceled.IsZero() {
		return fmt.Errorf("restaurant %s has been canceled", a.name)
	}

	if !a.scheduled.IsZero() {
		return fmt.Errorf(
			"restaurant %s is already scheduled for %s",
			a.name, a.scheduled.Format("2006-01-02"))
	}

	a.Root().Apply(&events.Scheduled{On: date})

	return nil
}

func (a *Restaurant) Cancel() error {
	var people []string

	if !a.canceled.IsZero() {
		return fmt.Errorf("%s already canceled", a.name)
	}

	for _, p := range a.subscriptions {
		people = append(people, p.person)
	}

	a.Root().Apply(&events.Canceled{
		Restaurant: a.name,
		People:     people,
		At:         time.Now()})

	return nil
}

func (a *Restaurant) Create(name, info string, menu ...string) error {
	if !a.created.IsZero() {
		return fmt.Errorf("restaurant %s is already created", a.name)
	}

	a.Root().Apply(&events.Created{
		Restaurant: name,
		Info:       info,
		Menu:       menu,
		At:         time.Now(),
	})

	return nil
}

func (a *Restaurant) TakeSnapshot() interface{} {
	return Snapshot{
		Version:       1,
		Name:          a.name,
		Info:          a.info,
		Menu:          a.menu,
		Subscriptions: a.subscriptions,
		Created:       a.created,
		Scheduled:     a.scheduled,
		Canceled:      a.canceled,
	}
}

func (a Restaurant) RestoreSnapshot(s interface{}) error {
	return nil
}

func handler(a *Restaurant) cqrs.DataHandler {
	return func(e interface{}) error {
		switch e := e.(type) {
		case *events.Created:
			a.name, a.info, a.menu = e.Restaurant, e.Info, e.Menu
			a.subscriptions = map[string]subscription{}
			a.created = e.At

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

		case *events.Rescheduled:
			a.scheduled = e.On

		default:
			return fmt.Errorf("event %s not handled", reflect.TypeOf(e))
		}

		return nil
	}
}
