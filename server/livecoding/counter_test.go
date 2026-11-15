package livecoding

import (
	"sync"
	"sync/atomic"
	"testing"
)

// §4.3 — counters shared across goroutines.
//
// The BROKEN variant is deliberately commented out: running it makes
// `go test -race` fail the whole suite (data race on c.n). Keep it as
// reference — reproducing the race locally is the lesson.
//
//	type brokenCounter struct{ n int }
//	func (c *brokenCounter) inc() { c.n++ }
//	func (c *brokenCounter) get() int { return c.n }

type mutexCounter struct {
	mu sync.Mutex
	n  int
}

func (c *mutexCounter) inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.n++
}

func (c *mutexCounter) get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.n
}

type atomicCounter struct {
	n int64
}

func (c *atomicCounter) inc()      { atomic.AddInt64(&c.n, 1) }
func (c *atomicCounter) get() int64 { return atomic.LoadInt64(&c.n) }

// hammer runs `inc` from `goroutines` goroutines, `perG` times each,
// then asserts the total — the WaitGroup owns goroutine completion.
func hammerCounter(t *testing.T, name string, inc func(), get func() int) {
	t.Helper()

	const goroutines, perG = 50, 1000
	var wg sync.WaitGroup

	wg.Add(goroutines) // add BEFORE launching, not inside the goroutine
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < perG; j++ {
				inc()
			}
		}()
	}
	wg.Wait()

	if got := get(); got != goroutines*perG {
		t.Errorf("%s: counter = %d, want %d (lost updates)", name, got, goroutines*perG)
	}
}

func TestCounterMutex(t *testing.T) {
	c := &mutexCounter{}
	hammerCounter(t, "mutex", c.inc, c.get)
}

func TestCounterAtomic(t *testing.T) {
	c := &atomicCounter{}
	hammerCounter(t, "atomic", c.inc, func() int { return int(c.get()) })
}
