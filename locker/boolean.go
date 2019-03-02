package locker

import "sync/atomic"

type Boolean struct{ v uint32 }

func (b *Boolean) Set(value bool) {
	if value {
		atomic.StoreUint32(&b.v, 1)
		return
	}

	atomic.StoreUint32(&b.v, 0)
}
func (b *Boolean) OK() bool { return atomic.LoadUint32(&b.v) != 0 }
