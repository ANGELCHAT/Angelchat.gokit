package scheduler

import (
	"time"
)

type Accuracy int

const (
	High Accuracy = iota
	Low
)

type Options struct {
	Workers int32
	Offsets map[time.Duration]Accuracy
}

type Option func(*Options)

func MaxWorkers(w int) Option {
	return func(o *Options) {
		o.Workers = int32(w)
	}
}

func SecondOffset(a Accuracy, ds ...time.Duration) Option {
	return func(o *Options) {
		o.Offsets = map[time.Duration]Accuracy{}
		for _, d := range ds {
			o.Offsets[d] = a
		}
	}
}

func newOptions(os ...Option) *Options {
	s := &Options{
		Workers: 1,
		Offsets: map[time.Duration]Accuracy{
			0: Low,
		},
	}
	for _, o := range os {
		o(s)
	}

	return s
}
