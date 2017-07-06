package scheduler

import "sync"

type Group struct {
	rs        []*Scheduler
	terminate chan<- struct{}
	done      <-chan struct{}
}

func NewGroup(rs []*Scheduler) *Group {

	t := make(chan struct{})
	d := make(chan struct{})

	go func() {
		var wg sync.WaitGroup
		<-t
		for _, j := range rs {
			wg.Add(1)
			go func(j *Scheduler) {
				j.Terminate()
				wg.Done()
			}(j)
		}
		wg.Wait()
		d <- struct{}{}
	}()

	return &Group{
		rs:        rs,
		done:      d,
		terminate: t,
	}
}

func (g *Group) Terminate() {
	g.terminate <- struct{}{}

	<-g.done
}

func (g *Group) Start() {
	for _, j := range g.rs {
		j.Schedule()
	}
}

func merge(cs ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	o := make(chan interface{}, len(cs))

	output := func(c <-chan interface{}) {
		for n := range c {
			o <- n
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(o)
	}()

	return o
}
