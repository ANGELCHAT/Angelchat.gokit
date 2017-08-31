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

func New() *aggregate {
	a := &aggregate{}
	a.root = cqrs.NewAggregate("tavern", handler(a))
	return a
}

func Load(id string) (*aggregate, error) {
	a := New()
	a.root.ID = cqrs.Identity(id)

	if err := service.Load(a.root); err != nil {
		return nil, err
	}

	return a, nil
}

func Save(a *aggregate) (string, error) {
	if err := service.Save(a.root); err != nil {
		return "", err
	}

	return a.root.ID.String(), nil
}
