package example

import (
	"fmt"

	"time"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/example/events"
	"github.com/sokool/gokit/cqrs/example/query"
)

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

var Query = query.New()

var service = cqrs.NewRepository(
	Factory,
	events.All,
	cqrs.EventHandler(Query.Listen),
)

func New() *Restaurant {
	return service.Aggregate().(*Restaurant)
}

func Load(id string) (*Restaurant, error) {
	a, err := service.Load(id)
	if err != nil {
		return nil, err
	}

	r, ok := a.(*Restaurant)
	if !ok {
		return nil, fmt.Errorf("wrong restaurant type")
	}

	return r, nil
}

func Save(a *Restaurant) error {
	return service.Save(a)
}

func Factory() (cqrs.Aggregate, cqrs.DataHandler) {
	r := &Restaurant{}
	return r, handler(r)
}
