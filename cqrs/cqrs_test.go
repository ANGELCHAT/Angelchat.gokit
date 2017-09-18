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
	log.Default = log.New(log.Levels(os.Stdout, os.Stdout, os.Stderr))
}

func TestRootID(t *testing.T) {
	// WHEN Restaurant aggregate is instantiated, without events
	// definitions
	repo := cqrs.NewRepository(example.Factory, nil)
	r1 := repo.Aggregate().(*example.Restaurant)

	// EXPECTS that ID is empty
	is.True(t, r1.Root().ID == "", "expects empty aggregate ID")

	// THEN Restaurant is saved,
	is.NotErr(t, repo.Save(r1))

	// EXPECTS that
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
	aggregate := cqrs.NewRepository(example.Factory, nil, cqrs.WithStorage(mem))

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
		cqrs.WithStorage(mem))

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
		cqrs.WithStorage(mem))

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
	aggregate := cqrs.NewRepository(example.Factory, events.All, cqrs.WithStorage(mem))

	r := aggregate.Aggregate().(*example.Restaurant)
	is.True(t, r.Root().Version == 0, "version 0 expected")
	is.NotErr(t, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.True(t, r.Root().Version == 0, "version 0 expected")
	is.NotErr(t, r.Subscribe("Tom", "Food"))
	is.True(t, r.Root().Version == 0, "version 0 expected")

	is.Ok(t, aggregate.Save(r))
	is.True(t, r.Root().Version == 2, "expected 2, got %d", r.Root().Version)

	is.Ok(t, r.Subscribe("Greg", "Burger"))
	is.True(t, r.Root().Version == 2, "expected 2, got %d", r.Root().Version)

	is.Ok(t, aggregate.Save(r))
	is.True(t, r.Root().Version == 3, "expected 3, got %d", r.Root().Version)

	r2, err := aggregate.Load(r.Root().ID)
	is.Ok(t, err)

	is.Ok(t, r.Subscribe("Albert", "Soup"))
	is.Ok(t, r.Subscribe("Mike", "Sandwitch"))
	is.Equal(t, uint64(3), r2.Root().Version)
	is.Ok(t, aggregate.Save(r))
	is.Equal(t, uint64(5), r.Root().Version)
}

//SCENARIO: Check aggregate transaction correctness
func TestTransactionCorrectness(t *testing.T) {
	//GIVEN I have Fresh Restaurant(r1) in version 3.
	mem := cqrs.NewMemoryStorage()
	repository := cqrs.NewRepository(example.Factory, events.All, cqrs.WithStorage(mem))
	r1 := repository.Aggregate().(*example.Restaurant)

	is.Ok(t, r1.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.Ok(t, r1.Subscribe("Tom", "Food"))
	is.Ok(t, r1.Subscribe("Tom", "Food2"))
	is.Ok(t, repository.Save(r1))
	is.Equal(t, uint64(3), r1.Root().Version)

	//WHEN I load that Restaurant (r2)
	a, err := repository.Load(r1.Root().ID)
	is.Ok(t, err)
	r2 := a.(*example.Restaurant)

	//AND I subscribe in r1 and r2 separately
	is.Ok(t, r1.Subscribe("Michel", "Y"))
	is.Ok(t, r2.Subscribe("Paula", "X"))

	//THEN I expect transaction error on second Restaurant while
	//storing both Restaurants
	is.Ok(t, repository.Save(r1))
	is.Err(t, repository.Save(r2), "transaction failed")
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
	is.NotErr(t, repo.Save(r))

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
	repo := cqrs.NewRepository(example.Factory, events.All,
		cqrs.WithStorage(store),
		cqrs.WithSnapshot(5, 500*time.Millisecond))

	// THEN I will crate Restaurant and assign 2 subscriptions, and wait 1 sec
	r := repo.Aggregate().(*example.Restaurant)
	r.Create("Restaurant A", "Description", "Meal A", "Meal B")
	r.Subscribe("Person#1", "A")
	r.Subscribe("Person#2", "D")
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first

	// I EXPECT Restaurant in version 3 and last snapshot in version 0
	snapVersion, _ := store.Snapshot(r.Root().ID)
	is.Equal(t, uint64(3), r.Root().Version)
	is.Equal(t, uint64(0), snapVersion)

	// THEN I add another 2 subscriptions and wait 1.5 sec
	r.Subscribe("Person#3", "A")
	r.Subscribe("Person#4", "D")
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first

	// I EXPECT restaurant in version 5 and snapshot in version 5.
	snapVersion, _ = store.Snapshot(r.Root().ID)
	is.Equal(t, uint64(5), r.Root().Version)
	is.Equal(t, uint64(5), snapVersion)

	// THEN I add another 4 Subscriptions
	for i := 5; i < 9; i++ {
		r.Subscribe(fmt.Sprintf("Person#%d", i), "A")
	}
	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first
	snapVersion, _ = store.Snapshot(r.Root().ID)

	// I EXPECT restaurant in version 9 and snapshot in version 5.
	is.Equal(t, uint64(9), r.Root().Version)
	is.Equal(t, uint64(5), snapVersion)

	// THEN I add another 3 more subscriptions and wait 1 sec
	r.Subscribe("Person#X", "Ax")
	r.Subscribe("Person#Y", "Dx")
	r.Subscribe("AL", "ad")

	is.Ok(t, repo.Save(r))
	time.Sleep(time.Second) // wait a while, to let snapshot run first
	snapVersion, _ = store.Snapshot(r.Root().ID)

	// I EXPECT restaurant in version 12 and snapshot in version 12.
	is.Equal(t, uint64(12), r.Root().Version)
	is.Equal(t, uint64(12), snapVersion)

}

func TestAggregateLoadFromLastSnapshot(t *testing.T) {
	// WHEN I tell repository to make a snapshot of Restaurant
	// every 0.1 second and every 2 events.
	store := cqrs.NewMemoryStorage()
	repo := cqrs.NewRepository(example.Factory, events.All,
		cqrs.WithStorage(store),
		cqrs.WithSnapshot(2, 100*time.Millisecond))

	// THEN I will crate Restaurant (r) and call 3 commands
	r := repo.Aggregate().(*example.Restaurant)
	is.Ok(t, r.Create("Restaurant A", "Description", "Meal A", "Meal B"))
	is.Ok(t, r.Subscribe("Person#1", "A"))
	is.Ok(t, r.Subscribe("Person#2", "D"))
	is.Ok(t, repo.Save(r))
	time.Sleep(500 * time.Millisecond) // wait a while, to let snapshot run first

	// THEN I will call 4 more commands
	is.Ok(t, r.Subscribe("Person#2", "A"))
	is.Ok(t, r.Subscribe("Person#2", "B"))
	is.Ok(t, r.Subscribe("Person#2", "C"))
	is.Ok(t, r.Subscribe("Person#2", "A"))
	is.Ok(t, repo.Save(r))
	time.Sleep(500 * time.Millisecond) // wait a while, to let snapshot run first

	//THEN I load that Restaurant (r2) again
	r2, err := repo.Load(r.Root().ID)
	is.Ok(t, err)

	// I EXPECT that last loaded aggregate from storage was
	// called with ID=r2.ID and from version=7
	is.Equal(t, r2.Root().ID, store.LastLoadID)
	is.Equal(t, uint64(7), store.LastLoadVersion)

	//THEN I load that Restaurant (r3) again
	a, err := repo.Load(r2.Root().ID)
	r3 := a.(*example.Restaurant)

	//I EXPECT last loaded Restaurant ID equal to previous
	//and last loaded version 7
	is.Ok(t, err)
	is.Equal(t, r2.Root().ID, store.LastLoadID)
	is.Equal(t, uint64(7), store.LastLoadVersion)

	// THEN I will call 4 more commands
	is.Ok(t, r3.Subscribe("Person#3", "A"))
	is.Ok(t, r3.Subscribe("Person#4", "B"))
	is.Ok(t, r3.Subscribe("Person#3", "B"))
	is.Ok(t, r3.Subscribe("Person#4", "A"))
	is.Ok(t, repo.Save(r3))

	//THEN I load that Restaurant again
	a, err = repo.Load(r3.Root().ID)
	r4 := a.(*example.Restaurant)

	//I EXPECT last loaded Restaurant ID equal to previous one
	//and last loaded version 7, but Restaurant in version 11
	is.Ok(t, err)
	is.Equal(t, r4.Root().ID, store.LastLoadID)
	is.Equal(t, uint64(7), store.LastLoadVersion)
	is.Equal(t, uint64(11), r4.Root().Version)

}

func TestRepositoryCache(t *testing.T) {
	// GIVEN Fresh Restaurant Repository with Enabled Cache
	store := cqrs.NewMemoryStorage()
	repo := cqrs.NewRepository(example.Factory, events.All,
		cqrs.WithStorage(store),
		cqrs.WithCache())

	// WHEN Restaurant(r1) in version 2 is created
	r1 := repo.Aggregate().(*example.Restaurant)
	is.Ok(t, r1.Create("a", "b", "c", "d", "e"))
	is.Ok(t, r1.Subscribe("dood", "meal"))

	// AND Restaurant (r1) is saved in repository
	is.Ok(t, repo.Save(r1))

	// WHEN loaded from Repository as r2
	a, err := repo.Load(r1.Root().ID)
	is.Ok(t, err)
	r2 := a.(*example.Restaurant)

	// EXPECTS that only 'Save' was called on Storage
	is.Equal(t, []string{"save"}, store.MethodCalls())

	// EXPECTS that r1 and r2 points to the same memory address.
	is.Equal(t, r1, r2)
	is.True(t, r1 == r2, "")

	// EXPECTS that Restaurant r2 was not loaded from Storage, but from
	// internal state(cache) of Repository
	is.Equal(t, uint64(0), store.LastLoadVersion)
	is.Equal(t, "", store.LastLoadID)

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
	repo := cqrs.NewRepository(example.Factory, events.All,
		cqrs.WithSnapshot(10, 500*time.Millisecond),
		cqrs.WithCache())

	r := repo.Aggregate().(*example.Restaurant)

	is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
	for i := 0; i < event; i++ {
		is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
	}

	is.Ok(b, repo.Save(r))

	for n := 0; n < b.N; n++ {
		a, err := repo.Load(r.Root().ID)
		rn := a.(*example.Restaurant)
		is.Ok(b, err)
		is.Ok(b, rn.Subscribe("Tom", "Papu"))
		//_, err = example.Save(a) // it's something wrong with this!
		//is.Ok(b, err)
	}
}
