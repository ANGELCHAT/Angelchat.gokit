package cqrs

import "github.com/sokool/gokit/log"

type Snapshotter struct {
	frequency  uint
	kind       string
	store      Store
	repo       *Repository
	serializer *serializer
	snapStruct structure
}

func (s *Snapshotter) Run() {
	as, err := s.store.Last(s.kind, s.frequency)
	if err != nil {
		log.Error("cqrs.snapshot.last", err)
		return
	}

	for _, a := range as {
		aggregate, err := s.repo.Load(a.ID)
		if err != nil {
			log.Error("cqrs.snapshot.aggregate.load", err)
			continue
		}
		//log.Info("cqrs.snap.test", aggregate.Root().String())
		snap, err := s.serializer.Marshal(s.snapStruct.Name, aggregate.TakeSnapshot())
		if err != nil {
			log.Error("cqrs.snapshot.marshal", err)
			continue
		}

		if err = s.store.Make(Snapshot{
			AggregateID: aggregate.Root().ID,
			Version:     aggregate.Root().Version,
			Data:        snap}); err != nil {

			log.Error("cqrs.snapshot.save", err)
			continue
		}

		log.Info("cqrs.snapshot.success", aggregate.Root().String())
	}
}

func (s *Snapshotter) Load(id string) (Aggregate, error) {

	version, data := s.repo.opts.Storage.Snapshot(id)
	aggregate, handler := s.repo.factory()
	root := newRoot(handler, s.repo.name)
	root.init(id, version)
	aggregate.Set(root)

	log.Info("cqrs.snapshot.load", "#%s v.%d", id[24:], version)
	if len(data) == 0 {
		return aggregate, nil
	}

	snapshot, err := s.serializer.Unmarshal(s.snapStruct.Name, data)
	if err != nil {
		return nil, err
	}

	if err := aggregate.RestoreSnapshot(snapshot); err != nil {
		return nil, err
	}

	return aggregate, nil
}

func NewSnapshotter(kind string, frequency uint, s Store, r *Repository) *Snapshotter {
	sStruct := r.Aggregate().TakeSnapshot()
	return &Snapshotter{
		frequency:  frequency,
		store:      s,
		repo:       r,
		kind:       kind,
		snapStruct: newStructure(sStruct),
		serializer: newSerializer(sStruct),
	}
}
