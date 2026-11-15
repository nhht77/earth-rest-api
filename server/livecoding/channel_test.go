package livecoding

import (
	"reflect"
	"testing"
	"time"
)

// §4.5 — FIFO semantics, close/range, and post-close behaviors.

func TestBufferedChannelFIFODrain(t *testing.T) {
	ch := make(chan int, 10)
	for i := 1; i <= 6; i++ {
		ch <- i // never blocks: 6 < capacity 10
	}
	close(ch)

	var got []int
	for v := range ch { // drains until closed AND empty, then ends cleanly
		got = append(got, v)
	}

	if !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5, 6}) {
		t.Errorf("FIFO order violated: %v", got)
	}
}

func TestSendAfterClosePanics(t *testing.T) {
	ch := make(chan int, 1)
	close(ch)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on send to closed channel")
		}
	}()

	ch <- 7
	t.Fatal("unreachable — the send must panic first")
}

func TestReceiveFromClosedAndEmpty(t *testing.T) {
	ch := make(chan int, 1)
	close(ch)

	v, ok := <-ch // zero value + ok=false, immediately, no blocking
	if ok {
		t.Error("ok should be false on closed+empty channel")
	}
	if v != 0 {
		t.Errorf("v = %d, want zero value 0", v)
	}
}

// Unbuffered = synchronous rendezvous: the send does NOT complete until a
// receiver takes the value. This is why unbuffered channels work as
// hand-off signals (close ≈ FIN on a socket).
func TestUnbufferedSendBlocksUntilReceive(t *testing.T) {
	ch := make(chan int) // capacity 0
	sent := make(chan struct{})

	go func() {
		ch <- 1
		close(sent)
	}()

	select {
	case <-sent:
		t.Fatal("unbuffered send completed without a receiver")
	case <-time.After(50 * time.Millisecond):
		// expected: still blocked
	}

	if got := <-ch; got != 1 {
		t.Errorf("received %d, want 1", got)
	}

	select {
	case <-sent:
		// send completed once the receive happened
	case <-time.After(time.Second):
		t.Fatal("send never completed after receive")
	}
}

// §1 — stack/queue on plain slices (channels are not LIFO).
type stack struct{ items []int }

func (s *stack) push(v int) { s.items = append(s.items, v) }
func (s *stack) pop() (int, bool) {
	if len(s.items) == 0 {
		return 0, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Queue with a head index: O(1) pop without re-slicing the whole slice.
type queue struct {
	items []int
	head  int
}

func (q *queue) push(v int) { q.items = append(q.items, v) }
func (q *queue) pop() (int, bool) {
	if q.head >= len(q.items) {
		return 0, false
	}
	v := q.items[q.head]
	q.head++
	if q.head == len(q.items) { // fully drained: reset storage
		q.items, q.head = nil, 0
	}
	return v, true
}

func TestStackIsLIFO(t *testing.T) {
	s := &stack{}
	s.push(1)
	s.push(2)
	s.push(3)

	for _, want := range []int{3, 2, 1} {
		if v, ok := s.pop(); !ok || v != want {
			t.Errorf("pop = (%d,%v), want (%d,true)", v, ok, want)
		}
	}
	if _, ok := s.pop(); ok {
		t.Error("empty stack pop should return ok=false")
	}
}

func TestQueueIsFIFO(t *testing.T) {
	q := &queue{}
	q.push(1)
	q.push(2)
	q.push(3)

	for _, want := range []int{1, 2, 3} {
		if v, ok := q.pop(); !ok || v != want {
			t.Errorf("pop = (%d,%v), want (%d,true)", v, ok, want)
		}
	}
	if _, ok := q.pop(); ok {
		t.Error("empty queue pop should return ok=false")
	}
}
