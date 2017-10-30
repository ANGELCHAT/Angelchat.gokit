package restaurant

import (
	"fmt"
	"time"

	"reflect"

	"encoding/json"

	"github.com/sokool/gokit/cqrs"
)

var Last *cqrs.Aggregate
var R *Restaurant
var Query = newQuery()

type subscription struct {
	Person string
	Meal   string
	On     time.Time
}

type Restaurant struct {
	ID            string
	Version       uint
	Name          string
	Info          string
	Menu          []string
	Subscriptions map[string]subscription

	Created   time.Time
	Scheduled time.Time
	Canceled  time.Time
}

func (a *Restaurant) String() string {
	b, _ := json.Marshal(a)
	return string(b)
}

func (a *Restaurant) TakeSnapshot() interface{} {
	return a
}

func (a *Restaurant) RestoreSnapshot(v interface{}) error {
	s, ok := v.(*Restaurant)
	if !ok {
		return fmt.Errorf("wront snapshot type[%s] in %s",
			reflect.TypeOf(s).String(),
			a.Name)
	}

	a.Name = s.Name
	a.Info = s.Info
	a.Menu = s.Menu
	a.Subscriptions = s.Subscriptions
	a.Created = s.Created
	a.Scheduled = s.Scheduled
	a.Canceled = s.Canceled

	return nil
}

func Factory() cqrs.FactoryFunc {
	return func(id string, version uint) *cqrs.Aggregate {
		r := &Restaurant{Subscriptions: make(map[string]subscription)}
		R = r
		Last = &cqrs.Aggregate{
			Name:    "restaurant",
			ID:      id,
			Version: version,
			Commands: map[cqrs.Command]cqrs.CommandHandler{
				&Create{}:     CreateHandler(r),
				&SelectMeal{}: selectMeal(r),
				&Cancel{}:     cancel(r),
				&Schedule{}:   schedule(r),
				&Reschedule{}: reschedule(r),
			},
			Events: map[interface{}]cqrs.EventHandler{
				&EventCreated{}:      created(r),
				&EventMealSelected{}: mealSelected(r),
				&EventMealChanged{}:  mealChanged(r),
				&EventScheduled{}:    scheduled(r),
				&EventRescheduled{}:  scheduled(r),
				&EventCanceled{}:     canceled(r),
			},
			TakeSnapshot:    r.TakeSnapshot,
			RestoreSnapshot: r.RestoreSnapshot,
		}

		return Last
	}
}

//var service = cqrs.NewEventSourced(
//	Factory,
//	cqrs.WithCache(),
//	cqrs.WithSnapshot(50, 10*time.Second),
//	cqrs.WithEventHandler(Query.Listen),
//	cqrs.WithEventStore(cqrs.NewMemoryStorage()),
//)
