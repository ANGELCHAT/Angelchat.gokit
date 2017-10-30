package cqrs_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/cqrs/restaurant"
	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
	"github.com/tonnerre/golang-pretty"
)

func init() {
	log.Default = log.New(
		log.Levels(os.Stdout, nil, os.Stderr),
		//log.Formatter(func(m log.Message) string {
		//	message := fmt.Sprintf(m.Text, m.Args...)
		//	data := time.Now().Format("05.0000")
		//
		//	return fmt.Sprintf("%s| %s: %s", data, m.Tag, message)
		//}),
	)
}

func eventRepository(f cqrs.FactoryFunc, r es.Storage) *es.EventStore {
	var el []interface{}
	for e := range f("", 0).Events {
		el = append(el, e)
	}

	return es.NewRepository(el,
		es.WithStore(r),
		es.WithListener(func(e es.event) {
			//log.Info("test.event.listener", "%+v", e)
		}))

}

func TestNoCommandHandler(t *testing.T) {
	factory := func(string, uint) *cqrs.Aggregate {
		return &cqrs.Aggregate{}
	}
	//GIVEN cqrs Aggregate without any command and event handlers.
	service := cqrs.NewEventSourced(factory, eventRepository(factory, es.NewMemStorage()))

	//WHEN sending restaurant.Create message
	r := service.Send(&restaurant.Create{})

	//THEN command handler not exists expected
	is.Err(t, r.Error, "command handler not exist")
	is.Equal(t, uint(0), r.Version)

}

func TestCommandHandling(t *testing.T) {
	//GIVEN Restaurant aggregate with Create command handler and
	//EventCreated registered.
	factory := func(id string, v uint) *cqrs.Aggregate {
		return &cqrs.Aggregate{
			Name: "test-aggregate",
			Commands: map[cqrs.Command]cqrs.CommandHandler{
				&restaurant.Create{}: restaurant.CreateHandler(&restaurant.Restaurant{}),
			},
			Events: map[interface{}]cqrs.EventHandler{
				&restaurant.EventCreated{}: nil,
			},
		}
	}

	repo := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, repo))

	//WHEN sending restaurant.Create command
	res := service.Send(&restaurant.Create{Name: "Test"})
	is.Ok(t, res.Error)

	//THEN aggregate is persisted  with one event.
	is.True(t, len(res.ID) > 0, "aggregate id expected")
	is.Equal(t, 1, repo.AggregatesCount())
	is.Equal(t, 1, repo.AggregatesEventsCount(res.ID))
	is.Equal(t, []string{"store"}, repo.MethodCalls())
}

func TestMultipleEventsInOneCommand(t *testing.T) {
	factory := func(string, uint) *cqrs.Aggregate {
		return &cqrs.Aggregate{
			Commands: map[cqrs.Command]cqrs.CommandHandler{
				&restaurant.Create{}: func(c cqrs.Command) ([]interface{}, error) {
					events := []interface{}{
						restaurant.EventCreated{},
						restaurant.EventCanceled{},
					}

					return events, nil
				},
			},
			Events: map[interface{}]cqrs.EventHandler{
				&restaurant.EventCreated{}:  nil,
				&restaurant.EventCanceled{}: nil,
			},
		}
	}

	storage := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, storage))

	res := service.Send(&restaurant.Create{})

	is.Ok(t, res.Error)
	is.Equal(t, 2, storage.AggregatesEventsCount(res.ID))
}

func TestEventsCountBySeparateCommands(t *testing.T) {
	// GIVEN Fresh Restaurant Service with Enabled Cache
	factory := restaurant.Factory()
	storage := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, storage))

	res1 := service.Send(&restaurant.Create{})
	is.Ok(t, res1.Error)
	is.True(t, len(res1.ID) > 0, "not empty id")

	res2 := service.Send(&restaurant.SelectMeal{}, res1.ID)
	is.Ok(t, res2.Error)
	is.Equal(t, res1.ID, res2.ID)

	is.Equal(t, res1.ID, res2.ID)
	is.Equal(t, res1.ID, storage.LastLoadID)
	is.Equal(t, uint(0), storage.LastLoadFromVersion)
	is.Equal(t, 1, storage.AggregatesCount())
	is.Equal(t, 2, storage.AggregatesEventsCount(res1.ID))
}

func TestAggregateVersion(t *testing.T) {
	factory := restaurant.Factory()
	repo := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, repo))

	is.Equal(t, uint(0), restaurant.Last.Version)

	res1 := service.Send(&restaurant.Create{Name: "Restaurant", Info: "Info", Menu: []string{"Meal A", "Meal B"}})
	is.Ok(t, res1.Error)
	is.Equal(t, uint(1), res1.Version)

	res2 := service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Food"}, res1.ID)
	is.Ok(t, res2.Error)
	is.Equal(t, uint(2), res2.Version)

	res3 := service.Send(&restaurant.SelectMeal{Person: "Greg", Meal: "Burger"}, res1.ID)
	is.Ok(t, res3.Error)
	is.Equal(t, uint(3), res3.Version)

	is.Equal(t, 1, repo.AggregatesCount())
	is.Equal(t, 3, repo.AggregatesEventsCount(res1.ID))

	res := service.Send(&restaurant.SelectMeal{Person: "Greg", Meal: "Burger"}, res1.ID)
	res = service.Send(&restaurant.SelectMeal{Person: "Greg", Meal: "Burger"}, res1.ID)
	res = service.Send(&restaurant.SelectMeal{Person: "Greg", Meal: "Burger"}, res1.ID)

	is.Equal(t, uint(6), res.Version)

}

//SCENARIO: Check aggregate transaction correctness
//func TestTransactionCorrectness(t *testing.T) {
//	//GIVEN I have Fresh Restaurant(r1) in version 3.
//	storage := es.NewMemStorage()
//	service := cqrs.NewEventSourced(restaurant.Factory(), cqrs.WithEventStore(storage))
//
//	var r cqrs.Response
//
//	r = service.Send(&restaurant.Create{Name: "Restaurant", Info: "Info", Menu: []string{"Meal A", "Meal B"}})
//
//	is.Ok(t, r.Error)
//	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Food"}, r.ID).Error)
//	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Other food"}, r.ID).Error)
//
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
//func TestSnapshotInGivenVersion(t *testing.T) {
//	// WHEN I tell service to make a snapshot of Restaurant every
//	// 5 versions and every 0.1 second
//	storage := es.NewMemStorage()
//	service := cqrs.NewEventSourced(restaurant.Factory(),
//		cqrs.WithEventStore(storage),
//		cqrs.WithSnapshot(5, 100*time.Millisecond))
//
//	// THEN I will crate Restaurant and assign 2 subscriptions, and wait 1 sec
//	id, err := service.Send(&restaurant.Create{
//		Name: "Restaurant A",
//		Info: "Description",
//		Menu: []string{"a", "b", "c"},
//	})
//	is.Ok(t, err)
//	service.Send(&restaurant.SelectMeal{Person: "Person#1", Meal: "A"}, id)
//	service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "D"}, id)
//	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first
//
//	// I EXPECT Restaurant in version 3 and last snapshot in version 0
//	snapVersion, _ := storage.Snapshot(id)
//	is.Equal(t, uint(3), restaurant.Recent.Version)
//	is.Equal(t, uint(0), snapVersion)
//
//	// THEN I add another 2 subscriptions and wait a while
//	service.Send(&restaurant.SelectMeal{Person: "Person#3", Meal: "A"}, id)
//	service.Send(&restaurant.SelectMeal{Person: "Person#4", Meal: "D"}, id)
//	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first
//
//	// I EXPECT restaurant in version 5 and snapshot in version 5.
//	snapVersion, _ = storage.Snapshot(id)
//	is.Equal(t, uint(5), restaurant.Recent.Version)
//	is.Equal(t, uint(5), snapVersion)
//
//	// THEN I add another 4 Subscriptions
//	for i := 5; i < 9; i++ {
//		service.Send(&restaurant.SelectMeal{Person: fmt.Sprintf("Person#%d", i), Meal: "A"}, id)
//	}
//	time.Sleep(150 * time.Millisecond) // wait a while, to let snapshot run first
//	snapVersion, _ = storage.Snapshot(id)
//
//	// I EXPECT restaurant in version 9 and snapshot in version 5.
//	is.Equal(t, uint(9), restaurant.Recent.Version)
//	is.Equal(t, uint(5), snapVersion)
//
//	// THEN I add another 3 more subscriptions and wait 1 sec
//	service.Send(&restaurant.SelectMeal{Person: "Person#X", Meal: "Ax"}, id)
//	service.Send(&restaurant.SelectMeal{Person: "Person#Y", Meal: "Dx"}, id)
//	service.Send(&restaurant.SelectMeal{Person: "AL", Meal: "ad"}, id)
//
//	time.Sleep(150 * time.Millisecond) // wait a while, to let snapshot run first
//	snapVersion, _ = storage.Snapshot(id)
//
//	// I EXPECT restaurant in version 12 and snapshot in version 12.
//	is.Equal(t, uint(12), restaurant.Recent.Version)
//	is.Equal(t, uint(12), snapVersion)
//
//}

func TestLoadFromLastSnapshot(t *testing.T) {
	// WHEN I tell service to make a snapshot of Restaurant
	// every 0.1 second and every 2 events.
	factory := restaurant.Factory()
	repo := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, repo),
		cqrs.WithSnapshot(2, 100*time.Millisecond))

	// THEN I will crate Restaurant and select 3 meals
	r := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Description",
		Menu: []string{"a", "b", "c"},
	})
	is.Ok(t, r.Error)
	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Person#1", Meal: "A"}, r.ID).Error)
	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "D"}, r.ID).Error)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first

	// THEN I will select 4 more meals
	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "A"}, r.ID).Error)
	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "B"}, r.ID).Error)
	is.Ok(t, service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "C"}, r.ID).Error)
	r1 := service.Send(&restaurant.SelectMeal{Person: "Person#2", Meal: "D"}, r.ID)
	is.Ok(t, r1.Error)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first

	// I EXPECT that last loaded aggregate from repo was
	// called with ID=r2.ID and from version=7
	is.Equal(t, r.ID, repo.LastLoadID)
	is.Equal(t, uint(6), repo.LastLoadFromVersion)
	is.Equal(t, uint(7), r1.Version)

	//// THEN I will select 4 more meals
	service.Send(&restaurant.SelectMeal{Person: "Person#3", Meal: "A"}, r.ID)
	service.Send(&restaurant.SelectMeal{Person: "Person#4", Meal: "B"}, r.ID)
	service.Send(&restaurant.SelectMeal{Person: "Person#3", Meal: "B"}, r.ID)
	time.Sleep(200 * time.Millisecond) // wait a while, to let snapshot run first
	service.Send(&restaurant.SelectMeal{Person: "Person#4", Meal: "A"}, r.ID)
	//
	////I EXPECT last loaded Restaurant ID equal to previous one
	////and last loaded version 7, but Restaurant in version 11
	//is.Equal(t, r.ID, repo.LastLoadID)
	//is.Equal(t, uint(10), repo.LastLoadVersion)
	//is.Equal(t, uint(11), restaurant.Recent.Version)

}

//func TestRepositoryCache(t *testing.T) {
//	// GIVEN Fresh Restaurant Service with Enabled Cache
//	storage := es.NewMemStorage()
//	service := cqrs.NewEventSourced(restaurant.Factory(),
//		cqrs.WithEventStore(storage),
//		cqrs.WithCache())
//
//	// WHEN Restaurant(r1) in version 2 is created
//	id, _ := service.Send(&restaurant.Create{
//		Name: "Restaurant A",
//		Info: "Description",
//		Menu: []string{"a", "b", "c"},
//	})
//	service.Send(&restaurant.SelectMeal{Person: "dood", Meal: "meal"}, id)
//
//	// EXPECTS that only 'Save' was called on Storage
//	is.Equal(t, []string{"save", "load", "save"}, storage.MethodCalls())
//
//	// EXPECTS that Restaurant  was not loaded from Storage, but from
//	// internal state(cache) of Service
//	is.Equal(t, uint(0), storage.LastLoadVersion)
//	is.Equal(t, "", storage.LastLoadID)
//}

func TestRestaurantState(t *testing.T) {
	factory := restaurant.Factory()
	repo := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, repo))

	is.Equal(t, uint(0), restaurant.Last.Version)

	r := service.Send(&restaurant.Create{
		Name: "Restaurant",
		Info: "Info",
		Menu: []string{"Meal A", "Meal B"}})
	is.Ok(t, r.Error)
	is.Equal(t, uint(1), restaurant.Last.Version)

	service.Send(&restaurant.Schedule{time.Now().AddDate(0, 0, 1)}, r.ID)
	service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Food"}, r.ID)
	service.Send(&restaurant.SelectMeal{Person: "Greg", Meal: "Burger"}, r.ID)
	service.Send(&restaurant.SelectMeal{Person: "Tom", Meal: "Bagel"}, r.ID)
	service.Send(&restaurant.Cancel{}, r.ID)

	pretty.Println(restaurant.Last.ID, int(restaurant.Last.Version))

}

func TestAggregateTransaction(t *testing.T) {
	factory := restaurant.Factory()
	repo := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, repo))

	r := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Info A",
		Menu: []string{"Meal A", "Meal B"}})
	is.Ok(t, r.Error)

	//id2, err := service.Send(&restaurant.Create{
	//	Name: "Restaurant B",
	//	Info: "Info B",
	//	Menu: []string{"Meal C", "Meal D"}})
	//is.Ok(t, err)

	mealA := &restaurant.SelectMeal{"Mike", "A", 100 * time.Millisecond}
	mealB := &restaurant.SelectMeal{"Tom", "B", 100 * time.Millisecond}
	mealC := &restaurant.SelectMeal{"Greg", "A", 100 * time.Millisecond}

	service.Send(mealA, r.ID)
	service.Send(mealB, r.ID)
	service.Send(mealC, r.ID)

	//log.Info("transaction", "\n%s\n%s", id1, id2)
}

//
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
	factory := restaurant.Factory()
	repo := es.NewMemStorage()
	service := cqrs.NewEventSourced(factory, eventRepository(factory, repo),
		cqrs.WithCache())

	r := service.Send(&restaurant.Create{
		Name: "Restaurant A",
		Info: "Description",
		Menu: []string{"a", "b", "c"},
	})
	service.Send(&restaurant.Schedule{time.Now().AddDate(0, 0, 1)}, r.ID)
	for i := 0; i < event; i++ {
		service.Send(&restaurant.SelectMeal{Person: fmt.Sprintf("Person #%d", i), Meal: "A"}, r.ID)
	}

	for n := 0; n < b.N; n++ {
		service.Send(&restaurant.SelectMeal{
			Person: "PersonIX",
			Meal:   "A"}, r.ID)
	}
}
