package es

import (
	"bytes"
	bin "encoding/binary"
	"fmt"
	"time"

	"github.com/alecthomas/binary"
	"github.com/boltdb/bolt"
	"github.com/sokool/gokit/log"
)

type storage struct {
	db *bolt.DB
}

func (s *storage) append(t Stream) (Provider, []Event, error) {
	var events []Event
	var stream Provider
	err := s.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(providers)
		if bp == nil {
			return fmt.Errorf("no %s bucket", providers)
		}

		// load or save Stream
		stream = Provider{ID: t.ID, Type: t.Name}
		if d := bp.Get([]byte(t.ID)); d == nil {
			if err := put(bp, []byte(stream.ID), &stream); err != nil {
				return err
			}
		} else if err := deserialize(d, &stream); err != nil {
			return err
		}

		be, err := tx.CreateBucketIfNotExists([]byte(stream.ID))
		if err != nil {
			return err
		}

		// store events with same time signature
		now := time.Now()
		for _, e := range t.Events {
			stream.Version++
			evt := Event{
				ID:      e.ID,
				Type:    e.Name,
				Data:    e.Data,
				Meta:    t.Meta,
				Version: stream.Version,
				Created: now,
			}

			if err := put(be, vtob(stream.Version), &evt); err != nil {
				return err
			}
			events = append(events, evt)
		}

		// store Stream with new version
		if err := put(bp, []byte(stream.ID), &stream); err != nil {
			return err
		}

		log.Debug("es.service.append", "v%d.%s with %d new events",
			stream.Version, stream.Type, len(events))

		return nil
	})

	return stream, events, err
}

func (s *storage) events(streamID string, from uint) ([]Event, error) {
	var events []Event

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(streamID))
		if b == nil {
			return fmt.Errorf("no events found for %s provider", streamID)
		}

		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			if from >= btov(k) {
				continue
			}

			e := Event{}
			if err := deserialize(v, &e); err != nil {
				return err
			}
			events = append(events, e)
		}

		return nil
	})

	//revers events
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}

	return events, err
}

func (s *storage) snapshot(d []byte, id string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bp := tx.Bucket(providers)
		if bp == nil {
			return fmt.Errorf("no providers bucket")
		}

		// load Provider
		pvr := Provider{}
		if d := bp.Get([]byte(id)); d == nil {
			return fmt.Errorf("provider %s not found", id)
		} else if err := deserialize(d, &pvr); err != nil {
			return err
		}

		bs := tx.Bucket(snapshots)
		if bs == nil {
			return fmt.Errorf("no %s bucket", snapshots)
		}

		snp := Snapshot{ID: pvr.ID, Version: pvr.Version, Data: d}
		if err := put(bs, []byte(snp.ID), &snp); err != nil {
			return err
		}

		return nil
	})
}

func (s *storage) providers(group string, older uint) ([]Provider, error) {
	var prs []Provider
	err := s.db.View(func(tx *bolt.Tx) error {
		bp := tx.Bucket(providers)
		if bp == nil {
			return fmt.Errorf("no providers bucket")
		}

		bs := tx.Bucket(snapshots)
		if bs == nil {
			return fmt.Errorf("no snapshots bucket")
		}

		c := bp.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			cs := bs.Cursor()
			for ks, vs := cs.Last(); ks != nil; ks, vs = cs.Prev() {
				if !bytes.Equal(k, ks) {
					continue
				}

				prv := Provider{}
				if err := deserialize(v, &prv); err != nil {
					return err
				}

				snp := Snapshot{}
				if err := deserialize(vs, &snp); err != nil {
					return err
				}

				if prv.ID != snp.ID || prv.Version < snp.Version+older {
					continue
				}
				fmt.Println(prv.Version)
				prv.Version = snp.Version
				prs = append(prs, prv)
			}

		}

		return nil
	})

	return prs, err
}

func newStorage(n string) (*storage, error) {
	db, err := bolt.Open(n, 0600, nil)
	if err != nil {
		return nil, err
	}

	// create buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(providers); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(events); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(snapshots); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &storage{
		db: db,
	}, nil
}

func put(b *bolt.Bucket, key []byte, value interface{}) error {
	if d, err := serialize(value); err != nil {
		return err
	} else if err = b.Put([]byte(key), d); err != nil {
		return err
	}
	return nil
}

func vtob(v uint) []byte {
	k := make([]byte, 4)
	bin.BigEndian.PutUint32(k, uint32(v))

	return k
}

func btov(k []byte) uint {
	return uint(bin.BigEndian.Uint32(k))
}

func deserialize(d []byte, v interface{}) error {
	return binary.Unmarshal(d, v)
}

func serialize(v interface{}) ([]byte, error) {
	return binary.Marshal(v)
}
