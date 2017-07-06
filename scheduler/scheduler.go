package scheduler

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/colchis/investor/internal/platform/log"
)

type Scheduler struct {
	opt  *Options
	max  int32
	task Task
	tmr  <-chan time.Time
	end  chan struct{}
	stop bool
}

func (p *Scheduler) Schedule() <-chan struct{} {
	p.stop = false
	go func() {
		var wg sync.WaitGroup
		for d := range p.tmr {
			if load(&p.max) == p.opt.Workers && p.opt.Workers > 0 {
				log.Error("SCHEDULER.error", "maximum %d number of workers", load(&p.max))
				continue
			}

			if p.stop {
				continue
			}

			add(&p.max)
			wg.Add(1)
			go func(d time.Time) {
				p.task()
				sub(&p.max)
				wg.Done()
			}(d)
		}
	}()

	return p.end
}

func (p *Scheduler) Terminate() {
	p.stop = true
	close(p.end)
	for {
		if load(&p.max) == 0 {
			break
		}
	}
}

func New(t Task, o ...Option) *Scheduler {
	s := newOptions(o...)
	e := make(chan struct{})

	ts := []<-chan time.Time{}
	for d, a := range s.Offsets {
		//if d >= time.Second || d < 0 {
		//	return nil, errors.New("offset out of range")
		//}

		if a == High {
			ts = append(ts, accurateHigh(d))
			continue
		}
		ts = append(ts, accurateLow(d))
	}

	tc := make(chan time.Time)
	var wg sync.WaitGroup
	for _, c := range ts {
		wg.Add(1)
		go func(c <-chan time.Time) {
			for {
				select {
				case n := <-c:
					tc <- n
				case <-e:
					wg.Done()
					return
				}
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(tc)
	}()

	return &Scheduler{
		opt:  s,
		task: t,
		tmr:  tc,
		end:  e,
	}
}

func load(w *int32) int32 {
	return atomic.LoadInt32(w)
}

func sub(w *int32) {
	atomic.StoreInt32(w, load(w)-1)
}

func add(w *int32) {
	atomic.AddInt32(w, 1)
}

func accurateHigh(d time.Duration) <-chan time.Time {
	var o time.Duration
	var s string
	log.Debug("SCHEDULER.offset-accuracy.high", "%s", d)
	for {
		s = s + "."
		n1 := time.Now()
		t := n1.Truncate(time.Second).Add(d + time.Second + o)
		<-time.NewTimer(t.Sub(n1)).C
		if d > 0 {
			x := time.NewTicker(d).C
			<-x
		}

		n2 := time.Now()
		o = t.Sub(n2)
		a := time.Duration(n2.Nanosecond()) - d

		if a <= 500*time.Microsecond && a > -time.Millisecond {
			return time.NewTicker(time.Second).C
		}
	}

}

func accurateLow(d time.Duration) <-chan time.Time {
	log.Debug("SCHEDULER.offset-accuracy.low", "%s", d)
	n1 := time.Now()
	t := n1.Truncate(time.Second).Add(d + time.Second)
	<-time.NewTimer(t.Sub(n1)).C
	//n2 := time.Now()
	//
	//fmt.Printf("low offset accuracy %v: %vms\n", d, n2.Format(".000")[1:])

	return time.NewTicker(time.Second).C
}
