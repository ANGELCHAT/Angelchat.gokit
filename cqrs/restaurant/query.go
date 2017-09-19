package restaurant

import (
	"time"

	"github.com/sokool/gokit/cqrs"
)

type Tavern struct {
	ID        int
	UUID      string
	Name      string
	Info      string
	Menu      []string
	CreatedAt time.Time
}

type Person struct {
	ID   int
	Name string
}

type query struct {
	tid     int
	pid     int
	taverns map[string]Tavern
	people  map[string]Person
}

func (q *query) Listen(a cqrs.CQRSAggregate, ce []cqrs.Event, es []interface{}) {
	for _, event := range es {
		switch e := event.(type) {
		case *EventCreated:
			if _, ok := q.taverns[a.String()]; ok {
				break
			}

			q.taverns[a.String()] = Tavern{
				ID:        q.tid,
				UUID:      a.ID,
				Name:      e.Restaurant,
				Info:      e.Info,
				Menu:      e.Menu,
				CreatedAt: e.At,
			}
			q.tid++
		case *EventMealSelected:
			if _, ok := q.people[e.Person]; ok {
				break
			}

			q.people[e.Person] = Person{
				ID:   q.pid,
				Name: e.Person,
			}
			q.pid++
		}
	}
}

func (q *query) Taverns() map[string]Tavern {
	return q.taverns
}

func (q *query) People() map[string]Person {
	return q.people
}

func newQuery() *query {
	return &query{
		taverns: map[string]Tavern{},
		people:  map[string]Person{},
	}
}
