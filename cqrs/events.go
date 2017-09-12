package cqrs

import (
	"time"

	"fmt"
)

type events struct {
	store      Store
	serializer *serializer
}

func (p *events) load(a Aggregate, from uint64) (int, error) {
	events, err := p.store.Events(from, a.Root().ID)
	if err != nil {
		return 0, err
	}

	for i, event := range events {
		e, err := p.serializer.Unmarshal(event.Type, event.Data)
		if err != nil {
			return i, err
		}

		if err := a.Root().handler(e); err != nil {
			return i, err
		}
	}

	return len(events), nil
}

func (p *events) changed(current Aggregate) error {
	var id = current.Root().ID
	var version = current.Root().Version

	if len(id) > 0 {
		stored, err := p.store.Load(id)
		if err != nil {
			return err
		}

		if stored.Version != version {
			return fmt.Errorf("%s.#%s transaction failed, current version is %d, but stored is %d",
				current.Root().Type, id[24:], stored.Version, version)
		}
	}

	return nil
}

// Takes events from Aggregate.
// Store aggregate and new events.
// Increase Aggregate version +len(root.events).
// Clear Aggregate events buffer
func (p *events) save(a Aggregate) (int, error) {
	var root *Root = a.Root()
	var events []Event
	var id string = root.ID
	var version uint64 = root.Version

	//todo lock?
	// check if stored Aggregate has not been changed.
	if err := p.changed(a); err != nil {
		return 0, err
	}

	// prepare events for storage
	for i, event := range root.events {
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
			Created: time.Now(),
			Version: version,
		})

	}

	// generate Aggregate ID if empty
	if len(id) == 0 {
		id = generateID()
	}

	//log.Info("events.store", "", root.Type)
	// store aggregate state and it's events
	v := CQRSAggregate{id, root.Type, version}
	if err := p.store.Save(v, events); err != nil {
		return 0, err
	}

	// modify Aggregate state
	root.ID = id
	root.Version = version
	root.events = []interface{}{}

	return len(events), nil
}
