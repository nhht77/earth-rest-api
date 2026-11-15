# Live Coding Prep — Go

> Interview scenarios where live coding likely hits. Each topic: what they
> test, the minimal pattern to have ready, one exercise to do in `gowolt`.

## 0. Problem → solution (the meta-skill)

The framing the interviewers stated: they watch *how you iterate*, not the
perfect first answer.

**Ready-made script for every live problem:**
1. Restate the problem in your own words; confirm assumptions out loud
2. Ask: input size? concurrent access? what happens on error/timeout?
3. Sketch the approach (data structure + flow) *before* writing code
4. Code the happy path; state what you're deferring ("error handling next")
5. Test with the edge cases you listed in step 2
6. If stuck: narrate the state — "we have X, we need Y, the gap is Z"

## 1. Stack / queue (as cache)

**What they test:** knowing when each fits (LIFO vs FIFO), implementing
without libs, and using them as building blocks (rate limiter, LRU).

**Minimal patterns:**
- Stack: slice + `append` / `s[len(s)-1]` + `s[:len(s)-1]`
- Queue: slice + head index (avoid O(n) re-slicing), or `container/list`
- As cache: LRU = `map[K]*list.Element` + doubly-linked list (evict tail)
- Rate limiter: token bucket (counter + refill ticker) or FIFO queue of
  timestamps

**Exercise (scratch repo, e.g. ~/dev/gowolt):** `lru.go` — LRU cache with `Get/Set`, 50-line cap,
table-driven tests incl. eviction order.

## 2. Goroutines — parallel vs async

**What they test:** when to fan out (parallel work) vs fire-and-forget
(async), and the sync primitives — plus *race awareness*.

**Minimal patterns:**
- Parallel: `sync.WaitGroup` + worker pool with a semaphore channel
- Async: producer/consumer over a channel; pipeline stages
- Shared state: `sync.Mutex` (or channels instead); always run tests with
  `-race`
- Errors from goroutines: `errgroup.Group` (golang.org/x/sync) — collects
  first error + context cancellation for free
- Never: goroutine without an owner (who stops it? who waits for it?)

**Exercise (scratch repo, e.g. ~/dev/gowolt):** fetch N URLs concurrently with worker pool (size 3),
collect results + first error, cancel remaining work on error. Test with
`httptest.Server` and `-race`.

**Real code reference:** `leasemanager/lease_manager.go` (atomic counters +
mutexes guarding shared state across consumers).

## 3. Context — cancellation and graceful shutdown

**What they test:** that cancellation *propagates* and that you can shut a
long-running app down cleanly. This is the #1 production-Go topic.

**Minimal patterns:**
- `ctx.Done()` in a `select` inside every long loop
- `context.WithTimeout` / `WithCancel` — always `defer cancel()`
- Pass ctx as **first parameter** through every layer (Neo v2 convention)
- Graceful shutdown sequence: catch SIGTERM → cancel context → HTTP server
  `Shutdown(ctx)` (drain in-flight) → close Kafka/DB (reverse order)
- The two-channel trick: `shutdownDone` or `sync.WaitGroup.Wait()` to know
  workers actually finished

**Real code reference (we've literally read it):**
`cmd/operations-execution/main.go` — `waitForSignal` turns SIGTERM into
`cancel()`; `app.Run` blocks on ctx; `Closers` run on exit.
operations-execution's whole API takes `ctx` first — including the logger.

**Exercise (scratch repo, e.g. ~/dev/gowolt):** a tiny server: `main` catches SIGINT/SIGTERM,
`http.Server.Shutdown(ctx)` with 5s timeout, a background worker stops
draining its channel on ctx.Done, logs "shutdown complete".

## 4. Practice backlog (my drafts — from hands-on exploration)

Each item: the concepts to nail, the gotchas I hit, exercise idea. Work
through one per sitting.

### 4.1 Values & mutation (arrays, structs, copies)
- [ ] Primitive array copy — arrays are *values* in Go: `b := a` is a full copy
- [ ] Struct array copy — copy is shallow; contained pointers/slices still shared!
- [ ] Struct copy — value receiver mutates a copy; pointer receiver mutates the original
- [ ] Slice vs array: a slice copy shares the backing array → mutations leak across "copies"
- Exercise: table-driven tests asserting what mutated and what didn't

### 4.2 Error checking
- [ ] `errors.New` + custom sentinel error types
- [ ] `errors.Is` — matching through `%w`-wrapped chains
- [ ] `errors.As` — extracting a typed error
- Exercise: build a 3-level error chain, assert `Is` + `As` behavior

### 4.3 Goroutines incrementing the same counter
- [ ] Unprotected increment → data race (prove with `go test -race`)
- [ ] Fix with `sync.Mutex`; alternative `atomic.AddInt64`
- [ ] `sync.WaitGroup` (Add/defer Done/Wait) to know when all finish
- [ ] `go func() {}()` — inline goroutine; who owns/stops it?
- Exercise: same counter, 3 versions (broken/mutex/atomic), each with a test

### 4.4 Context cancel & done (`fetchReport`)
- [ ] Exhaustive select: both `timer.C` and `ctx.Done()` cases — commenting one
      out makes the test vacuous (the timer case "passes" after 2s regardless)
- [ ] **`go vet` catches leaked cancels**: `ctx, _ := WithTimeout(...)` without
      `defer cancel()` → vet error "the cancel function is not used on all paths"
- [ ] `context.DeadlineExceeded` vs `context.Canceled` — which one fires?
- Already partially done in `pkg/logging/logging_test.go` (fetchReport exercise)

### 4.5 FIFO channel (buffered) — the socket analogy
- [ ] Buffered vs unbuffered: queue vs synchronous rendezvous
- [ ] `close` + `range` drain semantics; send-after-close panics; receive from
      closed+empty gives zero value with `ok == false`
- [ ] TCP/UDP analogy: channel ≈ socket buffer; close ≈ FIN
- ⚠️ gotcha hit: function named `testFIFO...` — tests MUST start with `Test`
- Exercise: assert the panic on post-close send + the `ok == false` receive

### 4.6 Expiring cache (single entry / whole list)
- [ ] Per-entry TTL: `map[K]entry{value, expiresAt}`; lazy expiry on read
- [ ] Whole-cache TTL: single timestamp; simplest correct approach
- [ ] Janitor goroutine for eager expiry — must own its stop channel (who stops it?)
- Exercise: expiring cache with `Get/Set`, expiry tests using injected clock
  (not `time.Sleep`) — clock injection is the interview-grade answer

## How to practice (low-energy-friendly)

- One exercise per sitting, 30–45 min, in `gowolt` — repo is tiny on purpose
- Narrate out loud while coding (the communication criterion)
- Every exercise: run with `-race`, then write 1-line "what surprised me"
- Log progress in `interview-fast-track.md` progress table
