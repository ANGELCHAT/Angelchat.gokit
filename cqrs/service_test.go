package cqrs_test

import (
	"testing"

	"github.com/sokool/gokit/cqrs"
	"github.com/sokool/gokit/cqrs/example"
	"github.com/sokool/gokit/cqrs/example/events"
	"github.com/sokool/gokit/test/is"
)

func TestEventRegistration(t *testing.T) {
	// Instantiate repository for restaurant without registered events.
	// Create aggregate (Restaurant) by calling Create command, and add
	// Burger Subscription for Tom. When aggregate is saved, expect error.

	service := cqrs.New(nil)

	restaurant := example.Restaurant()
	restaurant.Create("McKensey!", "Fine burgers!")
	restaurant.Subscribe("Tom", "Burger")

	is.True(t, service.Save(restaurant.Root) != nil, "expects not registered events error")
	is.True(t, restaurant.Root.ID.String() == "", "expects empty ID")

	service = cqrs.New([]interface{}{
		events.Created{},
		events.MealSelected{},
	})

	restaurant = example.Restaurant()
	restaurant.Create("McKensey!", "Fine burgers!")
	restaurant.Subscribe("Tom", "Burger")

	is.True(t, service.Save(restaurant.Root) == nil, "not expected error while saving aggregate")
	is.True(t, restaurant.Root.ID.String() != "", "not expected empty aggregate ID")

}
