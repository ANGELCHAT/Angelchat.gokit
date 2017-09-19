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

type Snapshot struct {
	Version uint

	Name string
	Info string
	Menu []string

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
	return Snapshot{
		Version:       1,
		Name:          a.Name,
		Info:          a.Info,
		Menu:          a.Menu,
		Subscriptions: a.Subscriptions,
		Created:       a.Created,
		Scheduled:     a.Scheduled,
		Canceled:      a.Canceled,
	}
}

func (a *Restaurant) RestoreSnapshot(v interface{}) error {
	s, ok := v.(*Snapshot)
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
		r := &Restaurant{ID: id, Version: version}
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
			Events: map[cqrs.Event2]cqrs.EventHandler{
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

//var service = cqrs.NewService(
//	Factory,
//	cqrs.WithCache(),
//	cqrs.WithSnapshot(50, 10*time.Second),
//	cqrs.WithEventHandler(Query.Listen),
//	cqrs.WithStorage(cqrs.NewMemoryStorage()),
//)
