package locker_test

import (
	"testing"
	"time"

	"github.com/livechat/gokit/locker"
	"github.com/livechat/gokit/log"
)

func action(l *locker.Key, key string, attempt int, d time.Duration) {
	log.Info("test.locking.came", "%s.#%d ", key, attempt)
	l.Lock(key)
	defer l.Unlock(key)

	time.Sleep(d)

	log.Info("test.locking.done", "%s.#%d ", key, attempt)
}

func TestLocker(t *testing.T) {
	l := locker.NewKey()

	go action(l, "A", 1, 100*time.Millisecond)
	go action(l, "A", 2, 100*time.Millisecond)
	go action(l, "A", 3, 100*time.Millisecond)
	go action(l, "A", 4, 100*time.Millisecond)
	go action(l, "A", 5, 100*time.Millisecond)
	go action(l, "A", 6, 100*time.Millisecond)

	time.Sleep(time.Millisecond * 20)
	go action(l, "B", 1, 10*time.Millisecond)
	go action(l, "B", 2, 10*time.Millisecond)
	go action(l, "B", 3, 10*time.Millisecond)
	go action(l, "B", 4, 10*time.Millisecond)
	go action(l, "B", 5, 10*time.Millisecond)
	go action(l, "B", 6, 10*time.Millisecond)

	//go action(l, "B", 1, 60*time.Millisecond)
	//go action(l, "B", 2, 80*time.Millisecond)
	//go action(l, "B", 3, 30*time.Millisecond)

	time.Sleep(time.Second)

}
