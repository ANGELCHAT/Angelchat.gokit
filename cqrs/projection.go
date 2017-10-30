package cqrs

type Projection struct {
	Events map[Event2]EventHandler
}
