package cqrs_test

import (
	"testing"

	"os"

	"time"

	"fmt"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/restaurant"
	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
	"github.com/tonnerre/golang-pretty"
)

func init() {
	log.Default = log.New(log.Levels(os.Stdout, os.Stdout, os.Stderr))
}

func TestNoCommandHandler(t *testing.T) {
	//GIVEN cqrs Aggregate without any command and event handlers.
	service := cqrs.NewService(func(string, uint) *cqrs.Aggregate {
		return &cqrs.Aggregate{}
	})

	//WHEN sending restaurant.Create message
	id, err := service.Send(&restaurant.Create{})

	//todo return instance error from CQRS and compare if command handler not exist returned.
	//THEN command handler not exists expected
	is.Err(t, err, "command handler not exist")
	is.Equal(t, "", id)
}

func TestCommandHandling(t *testing.T) {
	//GIVEN Restaurant aggregate with Create command handler and
	//EventCreated registered.
	r := &restaurant.Restaurant{}
	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(func(id string, v uint) *cqrs.Aggregate {
		return &cqrs.Aggregate{
			Commands: map[cqrs.Command]cqrs.CommandHandler{
				&restaurant.Create{}: restaurant.CreateHandler(r),
			},
			Events: map[cqrs.Event2]cqrs.EventHandler{
				&restaurant.EventCreated{}: nil,
			},
		}
	}, cqrs.WithStorage(storage))

	//WHEN sending restaurant.Create message
	id, err := service.Send(&restaurant.Create{Name: "Test"})
	is.Ok(t, err)

	//THEN aggregate is persisted  with one event.
	is.True(t, len(id) > 0, "aggregate id expected")
	is.Equal(t, 1, storage.AggregatesCount())
	is.Equal(t, 1, storage.AggregatesEventsCount(id))
	is.Equal(t, []string{"save"}, storage.MethodCalls())
}

func TestAggregateAndEventsAppearanceInStorage(t *testing.T) {
	// GIVEN Fresh Restaurant Service with Enabled Cache

	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(restaurant.Factory(), cqrs.WithStorage(storage))

	c1 := &restaurant.Create{
		Name: "Restauracja",
		Info: "Informacjr na temat tej świtnej restauracji.",
		Menu: []string{"a", "b", "c"},
	}

	id1, err := service.Send(c1)
	is.Ok(t, err)
	is.True(t, len(id1) > 0, "not empty id")

	c2 := &restaurant.SelectMeal{
		Person: "Tom",
		Meal:   "Mmmmmmhhhmmmm...",
	}
	id2, err := service.Send(c2, id1)

	is.Ok(t, err)
	is.Equal(t, id1, id2)
	is.Equal(t, 1, storage.AggregatesCount())
	is.Equal(t, 2, storage.AggregatesEventsCount(id1))
}

func TestAggregateVersion(t *testing.T) {
	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(restaurant.Factory(), cqrs.WithStorage(storage))

	is.Equal(t, uint(0), restaurant.Last.Version)

	id, err := service.Send(&restaurant.Create{
		Name: "Restaurant",
		Info: "Info",
		Menu: []string{"Meal A", "Meal B"}})
	is.Ok(t, err)
	is.Equal(t, uint(1), restaurant.Last.Version)

	_, err = service.Send(&restaurant.SelectMeal{
		Person: "Tom",
		Meal:   "Food",
	}, id)
	is.Ok(t, err)
	is.Equal(t, uint(2), restaurant.Last.Version)

	_, err = service.Send(&restaurant.SelectMeal{
		Person: "Greg",
		Meal:   "Burger",
	}, id)
	is.Ok(t, err)
	is.Equal(t, uint(3), restaurant.Last.Version)

	is.Equal(t, 1, storage.AggregatesCount())
	is.Equal(t, 3, storage.AggregatesEventsCount(id))

}

//SCENARIO: Check aggregate transaction correctness
//func TestTransactionCorrectness(t *testing.T) {
//	//GIVEN I have Fresh Restaurant(r1) in version 3.
//	storage := cqrs.NewMemoryStorage()
//	service := cqrs.NewService(restaurant.Factory(), cqrs.WithStorage(storage))
//
//
//	is.Ok(t, r1.Create("Restaurant", "Info", "Meal A", "Meal B"))
//	is.Ok(t, r1.Subscribe("Tom", "Food"))
//	is.Ok(t, r1.Subscribe("Tom", "Food2"))
//	is.Ok(t, repository.save(r1))
//	is.Equal(t, uint64(3), r1.Root().Version)
//
//	//WHEN I load that Restaurant (r2)
//	a, err := repository.load(r1.Root().ID)
//	is.Ok(t, err)
//	r2 := a.(*example.Restaurant)
//
//	//AND I subscribe in r1 and r2 separately
//	is.Ok(t, r1.Subscribe("Michel", "Y"))
//	is.Ok(t, r2.Subscribe("Paula", "X"))
//
//	//THEN I expect transaction error on second Restaurant while
//	//storing both Restaurants
//	is.Ok(t, repository.save(r1))
//	is.Err(t, repository.save(r2), "transaction failed")
//}

//
//func TestEventHandling(t *testing.T) {
//	var result, expected string
//
//	handler := func(a cqrs.CQRSAggregate, es []cqrs.Event, ds []interface{}) {
//		for _, event := range ds {
//			switch e := event.(type) {
//			case *events.EventCreated:
//				result += e.Restaurant
//			case *events.EventScheduled:
//				result += e.On.Format("2006-01-02")
//			case *events.EventMealSelected:
//				result += e.Person + e.Meal
//			}
//		}
//	}
//
//	repo := cqrs.NewRepository(example.Factory, events.All, cqrs.WithEventHandler(handler))
//
//	r := repo.Aggregate().(*example.Restaurant)
//	r.Create("Tavern", "description", "a", "b", "c")
//	r.Schedule(time.Now().AddDate(0, 0, 1))
//	r.Subscribe("Tom", "Food A")
//	r.Subscribe("Greg", "Food B")
//	r.Subscribe("Janie", "Food C")
//	is.NotErr(t, repo.save(r))
//
//	expected = "Tavern" +
//		time.Now().AddDate(0, 0, 1).Format("2006-01-02") +
//		"TomFood A" +
//		"GregFood B" +
//		"JanieFood C"
//
//	is.Equal(t, expected, result)
//
//}
//
func TestSnapshotInGivenVersion(t *testing.T) {
	// WHEN I tell service to make a snapshot of Restaurant every
	// 5 versions and every 0.1 second
	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(restaurant.Factory(),
		cqrs.WithStorage(storage),
		cqrs.WithSnapshot(5, 100*time.Millisecond))

	// THEN I will crate Restaurant and assign 2 subscriptions, and wait 1 sec
	id, err := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Description",
		Menu: []string{"a", "b", "c"},
	})
	is.Ok(t, err)
	service.Send(&restaurant.SelectMeal{Person: "Person#1", Meal: "A"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "D"}, id)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first

	// I EXPECT Restaurant in version 3 and last snapshot in version 0
	snapVersion, _ := storage.Snapshot(id)
	is.Equal(t, uint(3), restaurant.Last.Version)
	is.Equal(t, uint(0), snapVersion)

	// THEN I add another 2 subscriptions and wait a while
	service.Send(&restaurant.SelectMeal{Person: "Person#3", Meal: "A"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#4", Meal: "D"}, id)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first

	// I EXPECT restaurant in version 5 and snapshot in version 5.
	snapVersion, _ = storage.Snapshot(id)
	is.Equal(t, uint(5), restaurant.Last.Version)
	is.Equal(t, uint(5), snapVersion)

	// THEN I add another 4 Subscriptions
	for i := 5; i < 9; i++ {
		service.Send(&restaurant.SelectMeal{Person: fmt.Sprintf("Person#%d", i), Meal: "A"}, id)
	}
	time.Sleep(150 * time.Millisecond) // wait a while, to let snapshot run first
	snapVersion, _ = storage.Snapshot(id)

	// I EXPECT restaurant in version 9 and snapshot in version 5.
	is.Equal(t, uint(9), restaurant.Last.Version)
	is.Equal(t, uint(5), snapVersion)

	// THEN I add another 3 more subscriptions and wait 1 sec
	service.Send(&restaurant.SelectMeal{Person: "Person#X", Meal: "Ax"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#Y", Meal: "Dx"}, id)
	service.Send(&restaurant.SelectMeal{Person: "AL", Meal: "ad"}, id)

	time.Sleep(150 * time.Millisecond) // wait a while, to let snapshot run first
	snapVersion, _ = storage.Snapshot(id)

	// I EXPECT restaurant in version 12 and snapshot in version 12.
	is.Equal(t, uint(12), restaurant.Last.Version)
	is.Equal(t, uint(12), snapVersion)

}

func TestAggregateLoadFromLastSnapshot(t *testing.T) {
	// WHEN I tell service to make a snapshot of Restaurant
	// every 0.1 second and every 2 events.
	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(restaurant.Factory(),
		cqrs.WithStorage(storage),
		cqrs.WithSnapshot(2, 100*time.Millisecond))

	// THEN I will crate Restaurant and select 3 meals
	id, err := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Description",
		Menu: []string{"a", "b", "c"},
	})
	is.Ok(t, err)
	service.Send(&restaurant.SelectMeal{Person: "Person#1", Meal: "A"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "D"}, id)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first

	// THEN I will select 4 more meals
	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "A"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "B"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "C"}, id)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first
	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "A"}, id)

	// I EXPECT that last loaded aggregate from storage was
	// called with ID=r2.ID and from version=7
	is.Equal(t, id, storage.LastLoadID)
	is.Equal(t, uint(6), storage.LastLoadVersion)
	is.Equal(t, uint(7), restaurant.Last.Version)

	// THEN I will select 4 more meals
	service.Send(&restaurant.SelectMeal{Person: "Person#3", Meal: "A"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#4", Meal: "B"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Person#3", Meal: "B"}, id)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first
	service.Send(&restaurant.SelectMeal{Person: "Person#4", Meal: "A"}, id)

	//I EXPECT last loaded Restaurant ID equal to previous one
	//and last loaded version 7, but Restaurant in version 11
	is.Ok(t, err)
	is.Equal(t, id, storage.LastLoadID)
	is.Equal(t, uint(10), storage.LastLoadVersion)
	is.Equal(t, uint(11), restaurant.Last.Version)

}

func TestRepositoryCache(t *testing.T) {
	// GIVEN Fresh Restaurant Service with Enabled Cache
	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(restaurant.Factory(),
		cqrs.WithStorage(storage),
		cqrs.WithCache())

	// WHEN Restaurant(r1) in version 2 is created
	id, _ := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Description",
		Menu: []string{"a", "b", "c"},
	})
	service.Send(&restaurant.SelectMeal{Person: "dood", Meal: "meal"}, id)

	// EXPECTS that only 'Save' was called on Storage
	is.Equal(t, []string{"save", "load", "save"}, storage.MethodCalls())

	// EXPECTS that Restaurant  was not loaded from Storage, but from
	// internal state(cache) of Service
	is.Equal(t, uint(0), storage.LastLoadVersion)
	is.Equal(t, "", storage.LastLoadID)
}

func TestRestaurantState(t *testing.T) {
	storage := cqrs.NewMemoryStorage()
	service := cqrs.NewService(restaurant.Factory(), cqrs.WithStorage(storage))

	is.Equal(t, uint(0), restaurant.Last.Version)

	id, err := service.Send(&restaurant.Create{
		Name: "Restaurant",
		Info: "Info",
		Menu: []string{"Meal A", "Meal B"}})
	is.Ok(t, err)
	is.Equal(t, uint(1), restaurant.Last.Version)

	service.Send(&restaurant.Schedule{time.Now().AddDate(0, 0, 1)}, id)
	service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Food"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Greg", Meal: "Burger"}, id)
	service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Bagel"}, id)
	service.Send(&restaurant.Cancel{}, id)

	pretty.Println(restaurant.Last.ID, int(restaurant.Last.Version))

}

//
//func BenchmarkTest(b *testing.B) {
//
//	for n := 0; n < b.N; n++ {
//
//	}
//}
//
//func BenchmarkEventsStorage1(b *testing.B)     { benchmarkEventsStorage(1, b) }
//func BenchmarkEventsStorage100(b *testing.B)   { benchmarkEventsStorage(100, b) }
//func BenchmarkEventsStorage1000(b *testing.B)  { benchmarkEventsStorage(1000, b) }
//func BenchmarkEventsStorage10000(b *testing.B) { benchmarkEventsStorage(10000, b) }
//func BenchmarkEventsStorage50000(b *testing.B) { benchmarkEventsStorage(50000, b) }
//
//func BenchmarkEventsLoading1(b *testing.B)     { benchmarkEventsLoading(1, b) }
func BenchmarkEventsLoading100(b *testing.B) { benchmarkEventsLoading(100, b) }

//func BenchmarkEventsLoading1000(b *testing.B) { benchmarkEventsLoading(1000, b) }

//func BenchmarkEventsLoading10000(b *testing.B) { benchmarkEventsLoading(10000, b) }
//func BenchmarkEventsLoading50000(b *testing.B) { benchmarkEventsLoading(50000, b) }

//
//func benchmarkEventsStorage(commands int, b *testing.B) {
//	log.Default = log.newQuery(log.Levels(nil, nil, os.Stderr))
//	aggregate := cqrs.NewRepository(example.Factory, events.All)
//
//	for n := 0; n < b.N; n++ {
//		r := aggregate.Aggregate().(*example.Restaurant)
//		is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
//		is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
//		for i := 0; i < commands-2; i++ {
//			is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
//		}
//
//		is.Ok(b, aggregate.save(r))
//	}
//}
//
func benchmarkEventsLoading(event int, b *testing.B) {
	log.Default = log.New(log.Levels(nil, nil, os.Stderr))
	service := cqrs.NewService(restaurant.Factory(),
		//cqrs.WithSnapshot(10, 100*time.Millisecond),
		cqrs.WithCache(),
	)
	id, _ := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Description",
		Menu: []string{"a", "b", "c"},
	})
	service.Send(&restaurant.Schedule{time.Now().AddDate(0, 0, 1)}, id)
	for i := 0; i < event; i++ {
		service.Send(&restaurant.SelectMeal{Person: fmt.Sprintf("Person #%d", i), Meal: "A"}, id)
	}

	for n := 0; n < b.N; n++ {
		service.Send(&restaurant.SelectMeal{Person: "PersonIX", Meal: "A"}, id)
	}
}
