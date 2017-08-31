package query

import (
	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/example/events"
)

type Tavern struct {
	ID   int
	UUID string
	Name string
	Info string
	Menu []string
}

type Person struct {
	ID   int
	Name string
}

type Query struct {
	tid     int
	pid     int
	taverns map[string]Tavern
	people  map[string]Person
}

func (q *Query) Listen(aggregate cqrs.Identity, r cqrs.Event, data interface{}) error {
	switch e := data.(type) {
	case *events.Created:
		if _, ok := q.taverns[aggregate.String()]; ok {
			break
		}

		q.taverns[aggregate.String()] = Tavern{
			ID:   q.tid,
			UUID: aggregate.String(),
			Name: e.Restaurant,
			Info: e.Info,
			Menu: e.Menu,
		}
		q.tid++
	case *events.MealSelected:
		if _, ok := q.people[e.Person]; ok {
			break
		}

		q.people[e.Person] = Person{
			ID:   q.pid,
			Name: e.Person,
		}
		q.pid++
	}

	return nil
}

func (q *Query) Taverns() map[string]Tavern {
	return q.taverns
}

func (q *Query) People() map[string]Person {
	return q.people
}

func New() *Query {
	return &Query{
		taverns: map[string]Tavern{},
		people:  map[string]Person{},
	}
}
