package events

import "time"

type (
	Canceled struct {
		Restaurant string
		People     []string
		At         time.Time
	}

	Scheduled struct {
		On time.Time
	}

	Rescheduled struct {
		On time.Time
	}

	Created struct {
		Restaurant string
		Info       string
		Menu       []string
		At         time.Time
	}

	MealSelected struct {
		Person string
		Meal   string
		At     time.Time
	}

	MealChanged struct {
		Person      string
		PreviewMeal string
		ActualMeal  string
		At          time.Time
	}
)

var List = []interface{}{
	&Created{},
	&Scheduled{},
	&Rescheduled{},
	&Canceled{},
	&MealChanged{},
	&MealSelected{},
}
