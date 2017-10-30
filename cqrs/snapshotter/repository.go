package snapshotter

import (
	"fmt"

	"reflect"

	"github.com/sokool/gokit/cqrs/platform"
)

type Snapshot struct {
	ID      string
	Name    string
	Data    []byte
	Version uint
}

type Repository struct {
	name       string
	serializer *platform.Serializer
	store      Storage
}

func (r *Repository) Take(id string, version uint, v interface{}) error {
	l, err := r.store.Load(id)
	if err != nil {
		return err
	}

	if version <= l.Version {
		return fmt.Errorf("version must be greater than already stored")
	}

	data, err := r.serializer.Marshal(r.name, v)
	if err != nil {
		return err
	}

	return r.store.Replace(Snapshot{
		ID:      id,
		Name:    r.name,
		Version: version,
		Data:    data,
	})

}

func (r *Repository) Restore(id string) (interface{}, uint, error) {
	l, err := r.store.Load(id)
	if err != nil {
		return nil, 0, err
	}
	if l.Version == 0 {
		return nil, 0, nil
	}

	v, err := r.serializer.Unmarshal(r.name, l.Data)
	if err != nil {
		return nil, 0, err
	}

	return v, l.Version, nil
}

func NewRepository(v interface{}, s Storage) *Repository {
	return &Repository{
		name:       reflect.TypeOf(v).Name(),
		serializer: platform.NewSerializer(v),
		store:      s,
	}
}

//Making a snapshot algorythym:
//todo: combine 1,2 into one?
//1. Load all Aggregates of given type where
//	  aggregate.version - last_snap.version > frequency
//2. Load last snapshot and restore it on Aggregate.
//3. Load all the Events from snap.version and process them on Aggregate
//4. Take snapshot of Aggregate
//5. Save snapshot of Aggregate with version = Aggregate.version
//type snapshotter struct {
//	// todo move them out of snapshotter? used only in run method
//	frequency uint
//	kind      string
//	factory   FactoryFunc
//
//	store *store.Storage
//	//serializer *serializer
//	snapshot structure
//}
//
//func (s *snapshotter) run(frequency time.Duration) {
//	log.Info("cqrs.snapshot", "starting %s snapshotter every %s and every %d version",
//		s.kind, frequency, s.frequency)
//
//	for range time.NewTicker(frequency).C {
//		log.Info("cqrs.snapshot", "running...")
//		//1. load aggregates... snapshooter
//		aggregates, err := s.store.Persistence.Aggregates(s.kind, s.frequency)
//		if err != nil {
//			log.Error("cqrs.snapshot.load-last", err)
//			return
//		}
//
//		for _, a := range aggregates {
//			aggregate := s.factory(a.ID, a.Version)
//
//			if err := s.take(aggregate); err != nil {
//				log.Error("cqrs.snapshot", err)
//			}
//		}
//	}
//
//	log.Info("cqrs.snapshot", "%s finished", s.kind)
//
//}
//
//func (s *snapshotter) take(a *Aggregate) error {
//	//2. load last snap and restore aggregate with given snapshot data
//	version, data := s.store.Persistence.Snapshot(a.ID)
//	if len(data) > 0 {
//		snapshot, err := s.serializer.Unmarshal(s.snapshot.Name, data)
//		if err != nil {
//			return err
//		}
//
//		if err := a.RestoreSnapshot(snapshot); err != nil {
//			return err
//		}
//	}
//
//	// 3. Load all the Events from snap.version and process them on Aggregate
//	sa := store.Aggregate{ID: a.ID, Version: a.Version, Type: a.Name}
//	events, err := s.store.LoadEvents(sa)
//	if err != nil {
//		return err
//	}
//
//	for _, e := range events {
//		a.apply(e)
//	}
//
//	// 4. Take snapshot of Aggregate
//	data, err = s.serializer.Marshal(s.snapshot.Name, a.TakeSnapshot())
//	if err != nil {
//		return err
//	}
//
//	// 5. Save snapshot of Aggregate with version = Aggregate.version
//	v := Snapshot{a.ID, data, a.Version}
//	if err = s.store.store.Make(v); err != nil {
//		return err
//	}
//
//	log.Info(
//		"cqrs.snapshot", "%s.#%s.v%d taken, rebuilded "+
//			"from .v%d, with %d processed events",
//		a.Name, a.ID[24:], a.Version, version, num)
//
//	return nil
//}
//
//func (s *snapshotter) restore(a *Aggregate) (uint, error) {
//	version, data := s.store.store.Snapshot(a.ID)
//
//	// we have snapshot, restore it!
//	if len(data) > 0 {
//		snapshot, err := s.serializer.Unmarshal(s.snapshot.Name, data)
//		if err != nil {
//			return 0, err
//		}
//
//		if err := a.RestoreSnapshot(snapshot); err != nil {
//			return 0, err
//		}
//	}
//
//	log.Info("cqrs.snapshot",
//		"%s.#%s.v%d restored from .v%d",
//		a.Name, a.ID[24:], a.Version, version)
//
//	return version, nil
//}
//
//
//func newSnapshotter(frequency uint, e *store.Storage, f FactoryFunc) *snapshotter {
//	a := f("", 0)
//	sStruct := a.TakeSnapshot()
//	return &snapshotter{
//		frequency: frequency,
//		store:     e,
//		factory:   f,
//		kind:      a.Name,
//		snapshot:  newStructure(sStruct),
//		//serializer: newSerializer(sStruct),
//	}
//}
