package cqrs

import (
	"encoding/json"
	"reflect"
	"time"

	"fmt"

	"github.com/google/uuid"
)

type serializer struct {
	object map[string]reflect.Type
}

type Record struct {
	ID        string
	Type      string
	CreatedAt time.Time
	Event     []byte
	Version   uint8
}

func (e Record) String() string {
	return fmt.Sprintf(
		"#%s...: %s at %s\n",
		e.ID[0:8], e.Type, e.CreatedAt.Format("2006-01-02 15:04"))
}

func (r *serializer) Marshal(e Event) (Record, error) {
	name, _ := eventInfo(e)
	if _, ok := r.object[name]; !ok {
		return Record{}, fmt.Errorf("event '%s' is not registerd", name)
	}

	data, err := json.Marshal(e)
	if err != nil {
		return Record{}, err
	}

	return Record{
		ID:        uuid.New().String(),
		Type:      name,
		CreatedAt: time.Now(),
		Event:     data,
	}, nil
}

func (r *serializer) Unmarshal(v Record) (Event, error) {
	t, ok := r.object[v.Type]
	if !ok {
		return nil, fmt.Errorf("event %s is not registerd", v.Type)
	}

	event := reflect.New(t).Interface()
	if err := json.Unmarshal(v.Event, event); err != nil {
		return nil, err
	}

	return event, nil
}

func eventInfo(e Event) (name string, kind reflect.Type) {

	t := reflect.TypeOf(e)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name = t.Name()
	kind = t

	return
}

func newSerializer(es ...Event) *serializer {
	o := map[string]reflect.Type{}
	for _, e := range es {
		n, t := eventInfo(e)
		o[n] = t
	}

	return &serializer{
		object: o,
	}
}
