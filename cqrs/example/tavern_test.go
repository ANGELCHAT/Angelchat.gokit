package example_test

import (
	"testing"

	"time"

	"os"

	"fmt"

	"github.com/sokool/gokit/cqrs/example"
	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
)

func TestAggregate(t *testing.T) {
	r1 := example.Restaurant()

	is.Ok(t, r1.Create("PasiBus", "burgers restaurant", "onion", "chilly"))
	is.Ok(t, r1.Schedule(time.Now().AddDate(0, 0, 3)))
	is.Ok(t, r1.Subscribe("Mike", "Onion Burger!"))
	is.Ok(t, r1.Subscribe("Zygmunt", "kalafiorowa"))
	id1, err := example.Save(r1)
	is.Ok(t, err)

	r2 := example.Restaurant()
	is.Ok(t, r2.Create("Zdrowe Gary", "polskie papu", "pomidorowa", "ogórkowa", "kalafiorowa"))
	is.Ok(t, r2.Subscribe("Mike", "pomidorowa"))
	is.Ok(t, r2.Subscribe("Greg", "ogórkowa"))
	id2, err := example.Save(r2)
	is.Ok(t, err)

	is.Ok(t, r1.Subscribe("Tom", "Chilly Boy"))
	is.Ok(t, r1.Subscribe("Mike", "Cheesburger"))
	is.Ok(t, r1.Cancel())
	id3, err := example.Save(r1)
	is.Ok(t, err)

	is.Equal(t, id1, id3)
	is.True(t, id1 != id2, "")

	_, err = example.Load(id1)
	is.Ok(t, err)

	zdrowe, err := example.Load(id2)
	is.Ok(t, err)

	zdrowe.Reschedule(time.Now().AddDate(0, 0, 5))
	zdrowe.Subscribe("Charlie", "gilowana pierś")
	_, err = example.Save(zdrowe)
	is.Ok(t, err)

	log.Info("", "")
	for _, ta := range example.Query.Taverns() {
		log.Info("example.query.tavern", "%+v", ta)
	}

	for _, ta := range example.Query.People() {
		log.Info("example.query.person", "%+v", ta)
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

func benchmarkEventsStorage(x int, b *testing.B) {
	log.Default = log.New(log.Levels(nil, nil, os.Stderr))
	for n := 0; n < b.N; n++ {
		r := example.Restaurant()
		is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
		is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
		for i := 0; i < x; i++ {
			is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
		}

		_, err := example.Save(r)
		is.Ok(b, err)
	}
}

func benchmarkEventsLoading(x int, b *testing.B) {
	log.Default = log.New(log.Levels(nil, nil, os.Stderr))
	r := example.Restaurant()
	is.Ok(b, r.Create("Restaurant", "Info", "Meal A", "Meal B"))
	is.Ok(b, r.Schedule(time.Now().AddDate(0, 0, 1)))
	for i := 0; i < x; i++ {
		is.Ok(b, r.Subscribe(fmt.Sprintf("Person #%d", i), "Meal"))
	}

	id, err := example.Save(r)
	is.Ok(b, err)

	for n := 0; n < b.N; n++ {
		_, err := example.Load(id)
		is.Ok(b, err)
	}
}
