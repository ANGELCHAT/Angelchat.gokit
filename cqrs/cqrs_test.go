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

func init() {
	log.Default = log.New(log.Levels(os.Stdout, nil, os.Stderr))
}

func TestRootID(t *testing.T) {
	// When restaurant aggregate is saved, new ID should be assigned
	// in restaurant root.

	repo := cqrs.NewRepository(example.Factory, nil)
	r1 := repo.Aggregate().(*example.Restaurant)

	is.True(t, r1.Root().ID == "", "expects empty aggregate ID")
	is.NotErr(t, repo.Save(r1))
	is.True(t, r1.Root().ID != "", "expects aggregate ID")

	// When restaurant is saving and error appears, then ID should not
	// be assigned in restaurant root aggregate.
	r2 := repo.Aggregate().(*example.Restaurant)
	is.NotErr(t, r2.Create("Name", "Info", "Meal"))
	is.NotErr(t, r2.Subscribe("Person", "My Meal!"))

	is.Err(t, repo.Save(r2), "while saving aggregate")
	is.True(t, r2.Root().ID == "", "aggregate ID should be empty")
}

func TestEventRegistration(t *testing.T) {
	// Instantiate repository for restaurant without registered events definitions.
	// Create (Restaurant) aggregate by calling Create command, and add
	// Burger Subscription for Tom. When aggregate is saved, expect error.

	repo := cqrs.NewRepository(example.Factory, nil)

	r := repo.Aggregate().(*example.Restaurant)
	r.Create("McKensey!", "Fine burgers!")
	r.Subscribe("Tom", "Burger")

	is.Err(t, repo.Save(r), "events are not registered")
	is.True(t, r.Root().ID == "", "expects empty ID")

	repo = cqrs.NewRepository(example.Factory, []interface{}{
		events.Created{},
		events.MealSelected{},
	})

	r = repo.Aggregate().(*example.Restaurant)
	r.Create("McKensey!", "Fine burgers!")
	r.Subscribe("Tom", "Burger")

	is.NotErr(t, repo.Save(r))
	is.True(t, r.Root().ID != "", "not expected empty aggregate ID")

}

func TestAggregateAndEventsAppearanceInStorage(t *testing.T) {
	// when I store restaurant without performing any command I expect that
	// aggregate appears in storage without any generated events.
	mem := cqrs.NewMemoryStorage()
	aggregate := cqrs.NewRepository(example.Factory, nil, cqrs.Storage(mem))

	r := aggregate.Aggregate().(*example.Restaurant)
	is.Ok(t, aggregate.Save(r))
	is.True(t, mem.AggregatesCount() == 1, "expected one aggregate in storage")
	is.True(t, mem.AggregatesEventsCount(r.Root().ID) == 0, "no events expected")

	// when I create restaurant with Create command, one event should
	// appear in storage.
	mem = cqrs.NewMemoryStorage()
	aggregate = cqrs.NewRepository(
		example.Factory,
		[]interface{}{events.Created{}},
		cqrs.Storage(mem))

	r = aggregate.Aggregate().(*example.Restaurant)
	is.Ok(t, r.Create("McKenzy Food", "Burgers"))
	is.Ok(t, aggregate.Save(r))
	is.True(t, mem.AggregatesCount() == 1, "")
	is.True(t, mem.AggregatesEventsCount(r.Root().ID) == 1, "one event expected")
}

func TestMultipleCommands(t *testing.T) {
	mem := cqrs.NewMemoryStorage()
	aggregate := cqrs.NewRepository(
		example.Factory,
		[]interface{}{events.Created{}, events.MealSelected{}},
		cqrs.Storage(mem))

	// when I send Create command twice, I expect error on second Create
	// command call. After that, only one Created event should appear in storage.
	r := aggregate.Aggregate().(*example.Restaurant)
	is.NotErr(t, r.Create("Restaurant", "Info"))
	is.Err(t, r.Create("Another", "another info"), "expects already created error")

	is.NotErr(t, aggregate.Save(r))
	is.True(t, mem.AggregatesCount() == 1, "expects only one aggregate")
	is.True(t, mem.AggregatesEventsCount(r.Root().ID) == 1, "expects only one event in storage")
}

func TestAggregateVersion(t *testing.T) {
	mem := cqrs.NewMemoryStorage()
	aggregate := cqrs.NewRepository(example.Factory, events.All, cqrs.Storage(mem))

	r := aggregate.Aggregate().(*example.Restaurant)
	is.True(t, r.Root().Version == 0, "version 0 expected")
	is.NotErr(t, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.True(t, r.Root().Version == 0, "version 0 expected")
	is.NotErr(t, r.Subscribe("Tom", "Food"))
	is.True(t, r.Root().Version == 0, "version 0 expected")

	is.NotErr(t, aggregate.Save(r))
	is.True(t, r.Root().Version == 2, "expected 2, got %d", r.Root().Version)
	is.NotErr(t, r.Subscribe("Greg", "Burger"))
	is.True(t, r.Root().Version == 2, "expected 2, got %d", r.Root().Version)

	is.NotErr(t, aggregate.Save(r))
	is.True(t, r.Root().Version == 3, "expected 3, got %d", r.Root().Version)

	r2, err := aggregate.Load(r.Root().ID)
	is.Ok(t, err)

	is.NotErr(t, r.Subscribe("Albert", "Soup"))
	is.NotErr(t, r.Subscribe("Mike", "Sandwitch"))
	is.Equal(t, uint64(3), r2.Root().Version)
	is.NotErr(t, aggregate.Save(r))
	is.Equal(t, uint64(5), r.Root().Version)
}

func TestEventHandling(t *testing.T) {
	var result, expected string

	handler := func(a cqrs.CQRSAggregate, es []cqrs.Event, ds []interface{}) {
		for _, event := range ds {
			switch e := event.(type) {
			case *events.Created:
				result += e.Restaurant
			case *events.Scheduled:
				result += e.On.Format("2006-01-02")
			case *events.MealSelected:
				result += e.Person + e.Meal
			}
		}
	}

	repo := cqrs.NewRepository(example.Factory, events.All, cqrs.EventHandler(handler))

	r := repo.Aggregate().(*example.Restaurant)
	r.Create("Tavern", "description", "a", "b", "c")
	r.Schedule(time.Now().AddDate(0, 0, 1))
	r.Subscribe("Tom", "Food A")
	r.Subscribe("Greg", "Food B")
	r.Subscribe("Janie", "Food C")
	repo.Save(r)

	expected = "Tavern" +
		time.Now().AddDate(0, 0, 1).Format("2006-01-02") +
		"TomFood A" +
		"GregFood B" +
		"JanieFood C"

	is.Equal(t, expected, result)

}

func TestSnapshotInGivenVersion(t *testing.T) {
	// WHEN I tell repository to make a snapshot of Restaurant every
	// 5 versions and every 0.5 second
	store := cqrs.NewMemoryStorage()
	repo := cqrs.NewRepository(example.Factory, events.All, cqrs.Storage(store))
	repo.Snapshotter(5, 500*time.Millisecond)

	// THEN I will crate Restaurant and assign 2 subscriptions, and wait 1 sec
	r := repo.Aggregate().(*example.Restaurant)
	r.Create("Restaurant A", "Description", "Meal A", "Meal B")
	r.Subscribe("Person#1", "A")
	r.Subscribe("Person#2", "D")
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first
	sv, _ := store.Snapshot(r.Root().ID)

	// I EXPECT Restaurant in version 3 and last snapshot in version 0
	is.Equal(t, uint64(3), r.Root().Version)
	is.Equal(t, uint64(0), sv)

	// THEN I add another 2 subscriptions and wait 1.5 sec
	r.Subscribe("Person#1", "A")
	r.Subscribe("Person#2", "D")
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first
	sv, _ = store.Snapshot(r.Root().ID)

	// I EXPECT restaurant in version 5 and snapshot in version 5.
	is.Equal(t, uint64(5), r.Root().Version)
	is.Equal(t, uint64(5), sv)

	// THEN I add another 4 Subscriptions
	for i := 0; i < 4; i++ {
		r.Subscribe(fmt.Sprintf("Person#%d", i), "A")
	}
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first
	sv, _ = store.Snapshot(r.Root().ID)

	// I EXPECT restaurant in version 9 and snapshot in version 5.
	is.Equal(t, uint64(9), r.Root().Version)
	is.Equal(t, uint64(5), sv)

}

func TestAggregateLoadFromLastSnapshot(t *testing.T) {
	// WHEN I tell repository to make a snapshot of Restaurant
	// every 0.5 second and every 2 events.
	store := cqrs.NewMemoryStorage()
	repo := cqrs.NewRepository(example.Factory, events.All, cqrs.Storage(store))
	repo.Snapshotter(2, 500*time.Millisecond)

	// THEN I will crate Restaurant and assign 5 subscriptions, waiting 1 sec
	r := repo.Aggregate().(*example.Restaurant)
	r.Create("Restaurant A", "Description", "Meal A", "Meal B")
	r.Subscribe("Person#1", "A")
	r.Subscribe("Person#2", "D")
	r.Subscribe("Person#2", "D")
	r.Subscribe("Person#2", "D")
	r.Subscribe("Person#2", "D")
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first

	//THEN I load that Restaurant again
	r2, err := repo.Load(r.Root().ID)
	is.Ok(t, err)

	// I EXPECT that last loaded aggregate from storage was
	// called with ID=r2.ID and from version=6
	is.Equal(t, r2.Root().ID, store.LastLoadID)
	is.Equal(t, uint64(6), store.LastLoadVersion)

}

func BenchmarkTest(b *testing.B) {

	for n := 0; n < b.N; n++ {

	}
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
	aggregate := cqrs.NewRepository(example.Factory, events.All)

	for n := 0; n < b.N; n++ {
		r := aggregate.Aggregate().(*example.Restaurant)
		is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
		is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
		for i := 0; i < commands-2; i++ {
			is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
		}

		is.Ok(b, aggregate.Save(r))
	}
}

func benchmarkEventsLoading(event int, b *testing.B) {
	log.Default = log.New(log.Levels(nil, nil, os.Stderr))
	aggregate := cqrs.NewRepository(example.Factory, events.All)
	r := aggregate.Aggregate().(*example.Restaurant)

	is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
	for i := 0; i < event; i++ {
		is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
	}

	is.Ok(b, aggregate.Save(r))

	for n := 0; n < b.N; n++ {
		a, err := aggregate.Load(r.Root().ID)
		rn := a.(*example.Restaurant)
		is.Ok(b, err)
		is.Ok(b, rn.Subscribe("Tom", "Papu"))
		//_, err = example.Save(a) // it's something wrong with this!
		//is.Ok(b, err)
	}
}
