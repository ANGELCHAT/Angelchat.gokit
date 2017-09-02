package cqrs_test

import (
	"testing"

	"fmt"
	"time"

	"os"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/example"
	"github.com/sokool/gokit/cqrs/example/events"
	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
)

func TestRootID(t *testing.T) {
	// When restaurant aggregate root is saved, new ID should be assigned
	// in restaurant root.
	repo := cqrs.New(nil)
	r1 := example.Restaurant()
	is.True(t, r1.Root.ID.String() == "", "expects empty aggregate ID")
	is.NotErr(t, repo.Save(r1.Root))
	is.True(t, r1.Root.ID.String() != "", "expects aggregate ID")

	// When restaurant is saving and error appears, then ID should not
	// be assigned in restaurant root repo.
	r2 := example.Restaurant()
	is.NotErr(t, r2.Create("Name", "Info", "Meal"))
	is.NotErr(t, r2.Subscribe("Person", "My Meal!"))

	is.Err(t, repo.Save(r2.Root), "while saving aggregate")
	is.True(t, r2.Root.ID.String() == "", "aggregate ID should be empty")
}

func TestEventRegistration(t *testing.T) {
	// Instantiate repository for restaurant without registered events.
	// Create (Restaurant) aggregate by calling Create command, and add
	// Burger Subscription for Tom. When aggregate is saved, expect error.

	repo := cqrs.New(nil)

	r := example.Restaurant()
	r.Create("McKensey!", "Fine burgers!")
	r.Subscribe("Tom", "Burger")

	is.Err(t, repo.Save(r.Root), "events are not registered")
	is.True(t, r.Root.ID.String() == "", "expects empty ID")

	repo = cqrs.New([]interface{}{
		events.Created{},
		events.MealSelected{},
	})

	r = example.Restaurant()
	r.Create("McKensey!", "Fine burgers!")
	r.Subscribe("Tom", "Burger")

	is.NotErr(t, repo.Save(r.Root))
	is.True(t, r.Root.ID.String() != "", "not expected empty aggregate ID")

}

func TestAggregateAndEventsAppearanceInStorage(t *testing.T) {
	// when I store restaurant without performing any command I expect that
	// aggregate appears in storage without any generated events.
	mem := cqrs.NewMemoryStorage()
	repo := cqrs.New(nil, cqrs.Storage(mem))

	r := example.Restaurant()
	is.Ok(t, repo.Save(r.Root))
	is.True(t, mem.AggregatesCount() == 1, "expected one aggregate in storage")
	is.True(t, mem.AggregatesEventsCount(r.Root.ID.String()) == 0, "no events expected")

	// when I create restaurant with Create command, one event should
	// appear in storage.
	mem = cqrs.NewMemoryStorage()
	repo = cqrs.New([]interface{}{events.Created{}},
		cqrs.Storage(mem))

	r = example.Restaurant()
	is.Ok(t, r.Create("McKenzy Food", "Burgers"))
	is.Ok(t, repo.Save(r.Root))
	is.True(t, mem.AggregatesCount() == 1, "")
	is.True(t, mem.AggregatesEventsCount(r.Root.ID.String()) == 1, "one event expected")

}

func TestMultipleCommands(t *testing.T) {
	mem := cqrs.NewMemoryStorage()
	repo := cqrs.New([]interface{}{events.Created{}, events.MealSelected{}},
		cqrs.Storage(mem))

	// when I send Create command twice, I expect error on second Create
	// command call. After that, only one Created event should appear in storage.
	r := example.Restaurant()
	is.NotErr(t, r.Create("Restaurant", "Info"))
	is.Err(t, r.Create("Another", "another info"), "expects already created error")

	is.NotErr(t, repo.Save(r.Root))
	is.True(t, mem.AggregatesCount() == 1, "expects only one aggregate")
	is.True(t, mem.AggregatesEventsCount(r.Root.ID.String()) == 1, "expects only one event in storage")
}

func TestAggregateVersion(t *testing.T) {
	mem := cqrs.NewMemoryStorage()
	repo := cqrs.New(events.List, cqrs.Storage(mem))

	r := example.Restaurant()
	is.True(t, r.Root.Version == 0, "version 0 expected")
	is.NotErr(t, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.True(t, r.Root.Version == 0, "version 0 expected")
	is.NotErr(t, r.Subscribe("Tom", "Food"))
	is.True(t, r.Root.Version == 0, "version 0 expected")

	is.NotErr(t, repo.Save(r.Root))
	is.True(t, r.Root.Version == 2, "expected 2, got %d", r.Root.Version)
	is.NotErr(t, r.Subscribe("Greg", "Burger"))
	is.True(t, r.Root.Version == 2, "expected 2, got %d", r.Root.Version)

	is.NotErr(t, repo.Save(r.Root))
	is.True(t, r.Root.Version == 3, "expected 3, got %d", r.Root.Version)

	var err error
	r2 := example.Restaurant()
	r2.Root, err = repo.Load(r.Root.ID.String(), example.Handler(r2))
	is.Ok(t, err)

	is.NotErr(t, r.Subscribe("Albert", "Soup"))
	is.NotErr(t, r.Subscribe("Mike", "Sandwitch"))
	is.True(t, r2.Root.Version == 3, "expected 3, got %d", r2.Root.Version)
	is.NotErr(t, repo.Save(r.Root))
	is.True(t, r.Root.Version == 5, "expected 3, got %d", r2.Root.Version)
}

func BenchmarkEventsStorage1(b *testing.B)     { benchmarkEventsStorage(1, b) }
func BenchmarkEventsStorage100(b *testing.B)   { benchmarkEventsStorage(100, b) }
func BenchmarkEventsStorage1000(b *testing.B)  { benchmarkEventsStorage(1000, b) }
func BenchmarkEventsStorage10000(b *testing.B) { benchmarkEventsStorage(10000, b) }
func BenchmarkEventsStorage50000(b *testing.B) { benchmarkEventsStorage(50000, b) }


func BenchmarkEventsLoading1(b *testing.B)     { benchmarkEventsLoading(1, b) }
func BenchmarkEventsLoading100(b *testing.B)   { benchmarkEventsLoading(100, b) }
func BenchmarkEventsLoading1000(b *testing.B)  { benchmarkEventsLoading(1000, b) }
func BenchmarkEventsLoading10000(b *testing.B) { benchmarkEventsLoading(10000, b) }
func BenchmarkEventsLoading50000(b *testing.B) { benchmarkEventsLoading(50000, b) }

func benchmarkEventsStorage(commands int, b *testing.B) {
	log.Default = log.New(log.Levels(nil, nil, os.Stderr))
	for n := 0; n < b.N; n++ {
		r := example.Restaurant()
		is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
		is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
		for i := 0; i < commands-2; i++ {
			is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
		}

		_, err := example.Save(r)
		is.Ok(b, err)
	}
}

func benchmarkEventsLoading(events int, b *testing.B) {
	log.Default = log.New(log.Levels(nil, nil, os.Stderr))
	r := example.Restaurant()
	is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
	for i := 0; i < events; i++ {
		is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
	}

	id, err := example.Save(r)
	is.Ok(b, err)

	for n := 0; n < b.N; n++ {
		//example.Load(id)
		a, err := example.Load(id)
		is.Ok(b, err)
		is.Ok(b, a.Subscribe("Tom", "Papu"))
		//_, err = example.Save(a) // it's something wrong with this!
		//is.Ok(b, err)
	}
}

