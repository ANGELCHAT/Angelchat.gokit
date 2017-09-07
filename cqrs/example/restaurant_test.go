package example_test

import (
	"testing"

	"time"

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
