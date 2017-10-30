package store

import (
	"fmt"
	"time"
)

type Persistence interface {
	//=================SNAPSHOT METHODS

	// Last is calculated by subtracting the last snapshot version
	// from the current version with a where clause that only returned the
	// aggregates with a difference greater than some number. This query
	// will return all of the Aggregates that a snapshot to be created.
	// The snapshotter would then iterate through this list of Aggregates
	// to create the snapshots (if using	multiple snapshotters the
	// competing consumer pattern works well here).
	Aggregates(kind string, vFrequency uint) ([]Aggregate, error)
	Make(s Snapshot) error
	// Snapshot returns last snapshoted version with respective data
	// it might return (0, nil) when snapshot is not stored
	// todo what about if aggregate not exists, at all?
	Snapshot(aggregate string) (uint, []byte)

	//=================EVENT METHODS
	Load(id string) (Aggregate, error)
	Save(Aggregate, []Event) error
	// load all aggregates and events from given version.
	Events(aggregate string, version uint) ([]Event, error)
}

type Snapshooter interface {
	Store(Snapshot) error
	Load(aggregate string) Snapshot
}

//type Stream chan interface{}

type Event struct {
	ID            string
	Type          string
	AggregateType string
	AggregateID   string
	Data          []byte
	Version       uint
	Created       time.Time
}

type Aggregate struct {
	ID      string
	Type    string
	Version uint
}

func (a *Aggregate) String() string {
	return fmt.Sprintf("#%s: v%d.%s",
		a.ID[24:], a.Version, a.Type)
}

type Snapshot struct {
	AggregateID string
	Data        []byte
	Version     uint
}
