package es_test

import (
	"testing"

	"os"

	"sync"

	"fmt"
	"time"

	"github.com/sokool/gokit/cqrs/es"
	"github.com/sokool/gokit/log"
	"github.com/sokool/gokit/test/is"
)

func init() {
	log.Default = log.New(
		log.Levels(os.Stdout, nil, os.Stderr),
		log.Formatter(func(m log.Message) string {
			message := fmt.Sprintf(m.Text, m.Args...)
			data := time.Now().Format("05.0000")

			return fmt.Sprintf("%s| %s: %s", data, m.Tag, message)
		}))

}

type Event struct {
	Info string
}

var events = []interface{}{
	Event{},
}

func TestA(t *testing.T) {
	s, err := es.NewService("db/bench-write")
	is.Ok(t, err)
	//
	//in := es.Stream{
	//	ID:    "x",
	//	Group: "aggregate",
	//}
	//for i := 0; i < 3; i++ {
	//	in.Data = append(in.Data, es.Data{Stream: "event", Data: []byte("{'test':1}")})
	//}
	//
	//is.Ok(t, s.Append(in))
	//is.Ok(t, s.Snapshot([]byte("serialized aggregate 2"), "x"))

	ps, err := s.Providers("aggregate", 5)
	is.Ok(t, err)
	for _, p := range ps {
		log.Info("test.snaps", "%+v", p)
	}

	//
	events, err := s.Events("x", 9999900)
	is.Ok(t, err)
	for _, e := range events {
		log.Info("test", "V.%d", e.Version)
	}

}

func benchmarkWriting(b *testing.B, id string, number int, concurrent int) {
	os.Remove("db/bench-write")
	s, err := es.NewService("db/bench-write")
	is.Ok(b, err)

	in := es.Stream{ID: id, Name: "aggregate"}
	for i := 0; i < number; i++ {
		in.Events = append(in.Events, es.Record{Name: "event", Data: []byte("{'test':1}")})
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		is.Ok(b, s.Append(in))
	}

}

func benchmarkLoading(b *testing.B, number int, from uint, concurrent int) {
	os.Remove("db/bench")
	s, err := es.NewService("db/bench")
	is.Ok(b, err)

	stream := es.Stream{
		ID:     "x",
		Name:   "aggregate",
		Events: []es.Record{},
	}

	for i := 0; i < number; i++ {
		stream.Events = append(stream.Events, es.Record{Name: "event", Data: []byte("{'test':1}")})
	}
	is.Ok(b, s.Append(stream))

	if concurrent > 0 {
		wg := sync.WaitGroup{}
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			for i := 0; i < concurrent; i++ {
				wg.Add(1)
				go func() {
					s.Events("x", from)
					wg.Done()
				}()
			}

			wg.Wait()
		}
	} else {
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			s.Events("x", from)
		}
	}
}

func BenchmarkLoading100EventsWith10Concurrent(b *testing.B) {
	benchmarkLoading(b, 100, 0, 10)
}

func BenchmarkLoading100EventsWith100Concurrent(b *testing.B) {
	benchmarkLoading(b, 100, 0, 100)
}

func BenchmarkLoading100EventsWith1kConcurrent(b *testing.B) {
	benchmarkLoading(b, 100, 0, 1000)
}

func BenchmarkLoading100EventsWith2kConcurrent(b *testing.B) {
	benchmarkLoading(b, 100, 0, 2000)
}

func BenchmarkLoading10EventsFrom0(b *testing.B) {
	benchmarkLoading(b, 10, 0, 0)
}

func BenchmarkLoading100EventsFrom0(b *testing.B) {
	benchmarkLoading(b, 100, 0, 0)
}

func BenchmarkLoading1kEventsFrom0(b *testing.B) {
	benchmarkLoading(b, 1000, 0, 0)
}

func BenchmarkLoading3kEventsFrom0(b *testing.B) {
	benchmarkLoading(b, 3000, 0, 0)
}

func BenchmarkWriting100Events(b *testing.B) {
	benchmarkWriting(b, "x", 100, 0)
}

func BenchmarkWriting1kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 1000, 0)
}

func BenchmarkWriting5kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 5000, 0)
}

func BenchmarkWriting10kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 10000, 0)
}

func BenchmarkWriting30kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 30000, 0)
}

func BenchmarkWriting50kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 50000, 0)
}

func BenchmarkWriting100kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 100000, 0)
}

func BenchmarkWriting300kEvents(b *testing.B) {
	benchmarkWriting(b, "x", 300000, 0)
}

//func TestDefaultRepository(t *testing.T) {
//	es.NewRepository(events)
//
//}
//
//func TestSavingEventsInvalidInput(t *testing.T) {
//	r := es.NewRepository(events)
//
//	v, err := r.Save("", "")
//	is.Equal(t, err, es.ErrNoAggregateIdentity)
//	is.Equal(t, uint(0), v)
//
//	v, err = r.Save("X8YSD31", "")
//	is.Equal(t, err, es.ErrWrongAggregateName)
//	is.Equal(t, uint(0), v)
//
//	v, err = r.Save("X8YSD31", "cool")
//	is.Equal(t, err, es.ErrEmptyEvents)
//
//}
//
//func TestEventVersion(t *testing.T) {
//	id := "a"
//	name := "aaa"
//
//	r := es.NewRepository(events)
//
//	v, err := r.Save(id, name, Event{"a"}, Event{"c"})
//	is.Ok(t, err)
//	is.Equal(t, uint(2), v)
//
//	v, err = r.Save(id, name, Event{"d"})
//	is.Ok(t, err)
//	is.Equal(t, uint(3), v)
//}
//
//func TestLoadedEventsOrder(t *testing.T) {
//	id := "a"
//	name := "aaa"
//	r := es.NewRepository(events)
//
//	v, err := r.Save(id, name,
//		Event{"a"},
//		Event{"c"},
//		Event{"d"},
//		Event{"b"})
//	is.Ok(t, err)
//	is.Equal(t, uint(4), v)
//
//	el, version, err := r.Load(id, 0)
//	is.Ok(t, err)
//	is.Equal(t, uint(4), version)
//
//	expected := []string{"a", "c", "d", "b"}
//	for i, v := range el {
//		e, ok := v.(*Event)
//		is.True(t, ok, "expected *Event")
//		is.Equal(t, expected[i], e.Info)
//	}
//}
//
//func TestLoadingFromVersion(t *testing.T) {
//	id := "a"
//	name := "aaa"
//
//	r := es.NewRepository(events)
//
//	v, err := r.Save(id, name,
//		Event{"a"},
//		Event{"c"},
//		Event{"d"},
//		Event{"b"})
//	is.Ok(t, err)
//	is.Equal(t, uint(4), v)
//
//	el, version, err := r.Load(id, 3)
//	is.Ok(t, err)
//	is.Equal(t, uint(4), version)
//
//	expected := []string{"d", "b"}
//	for i, v := range el {
//		e, ok := v.(*Event)
//		is.True(t, ok, "expected *Event")
//		is.Equal(t, expected[i], e.Info)
//	}
//
//	if err := es.Do(10); err == es.ErrTooBig {
//		//zabawa z obsługą
//	}
//}
