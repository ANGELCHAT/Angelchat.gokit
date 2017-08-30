package cqrs

import "time"

type Event interface {
	Aggregate(id string)
	CreatedAt(time.Time)
}

type Model struct {
	ID      string
	Created time.Time
}

func (m *Model) Aggregate(id string) {
	m.ID = id
}

func (m *Model) CreatedAt(d time.Time) {
	m.Created = d
}

