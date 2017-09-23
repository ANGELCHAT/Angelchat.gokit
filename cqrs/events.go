package cqrs

import (
	"time"

	"fmt"
)

type events struct {
	store      Store
	serializer *serializer
}

func (p *events) load(a *Aggregate, from uint) (int, error) {
	events, err := p.store.Events(from, a.ID)
	if err != nil {
		return 0, err
	}

	for i, event := range events {
		e, err := p.serializer.Unmarshal(event.Type, event.Data)
		if err != nil {
			return i, err
		}

		a.apply(e)
	}

	return len(events), nil
}

func (p *events) changed(current *Aggregate) error {
	var id = current.ID
	var version = current.Version

	if len(id) > 0 {
		stored, err := p.store.Load(id)
		if err != nil {
			return err
		}

		if uint(stored.Version) != version {
			return fmt.Errorf("%s.#%s transaction failed, current version is %d, but stored is %d",
				current.Name, id[24:], version, stored.Version)
		}
	}

	return nil
}

// Takes events from Aggregate.
// Store aggregate and new events.
// Increase Aggregate version +len(root.events).
// Clear Aggregate events buffer
func (p *events) save(a *Aggregate) (int, error) {
	var events []Event
	var date = time.Now()
	var version = a.Version
	var id = a.ID

	// check if stored Aggregate has not been changed.
	if err := p.changed(a); err != nil {
		return 0, err
	}

	// prepare events for storage
	for i, event := range a.events {
		version++
		name := newStructure(event).Name
		data, err := p.serializer.Marshal(name, event)
		if err != nil {
			return i, err
		}

		events = append(events, Event{
			ID:      generateID(),
			Type:    name,
			Data:    data,
			Created: date, // each event in transaction get the same date
			Version: version,
		})

	}

	// generate Aggregate ID if empty
	if len(a.ID) == 0 {
		id = generateID()
	}

	//log.Info("events.store", "", root.Type)
	// store aggregate state and it's events
	v := CQRSAggregate{id, a.Name, version}
	if err := p.store.Save(v, events); err != nil {
		return 0, err
	}

	// modify Aggregate state
	a.ID = id
	a.Version = version
	a.events = []Event2{}

	return len(events), nil
}
