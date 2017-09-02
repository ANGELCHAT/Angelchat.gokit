package example

import (
	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/example/events"
	"github.com/sokool/gokit/cqrs/example/query"
)

var Query = query.New()

var service = cqrs.New(
	events.List,
	cqrs.EventHandler(Query.Listen),
)

func Restaurant() *restaurant {
	a := &restaurant{}
	a.Root = cqrs.NewAggregate("restaurant", Handler(a))
	return a
}

func Load(id string) (*restaurant, error) {
	var err error
	a := Restaurant()
	if a.Root, err = service.Load(id, Handler(a)); err != nil {
		return nil, err
	}

	return a, nil
}

func Save(a *restaurant) (string, error) {
	if err := service.Save(a.Root); err != nil {
		return "", err
	}

	return a.Root.ID.String(), nil
}
