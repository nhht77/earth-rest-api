# Interview prep — links by topic

Created: 22.11.2026. Companion to `service-evolution-plan.md` — each link maps to
work actually done in this repo (see the coverage table below).

---

## Coverage audit: interview criterion → evidence → related links

| Interview criterion | What you've done (evidence you can describe) | Confidence | Related links |
|---|---|---|---|
| **Maintainability** | Split `pkg/v1` → `pkg/{city,country,continent,user}`; package names match dirs; moved `pkg/` to root (AANM layout); SetupFn ordered setup; the `pkg/mhttp → pkg/handlers` rename with all 7 importers | High | [Practical Go (Cheney)](https://dave.cheney.net/practical-go/presentations/qcon-china.html) · [Uber Go Style Guide](https://github.com/uber-go/guide) · [Effective Go](https://go.dev/doc/effective_go) |
| **Readability** | Phases 1–5 logging refactor; consistent error classification; `-- condition:` smell identified and removed; doc comments as decisions | High | [Practical Go (Cheney)](https://dave.cheney.net/practical-go/presentations/qcon-china.html) · [Go Proverbs](https://go-proverbs.github.io) · [Uber Go Style Guide](https://github.com/uber-go/guide) |
| **Testability** | `httptest.NewRecorder` vs `NewServer(RunHTTP(h))` vs IT-behind-build-tag layering; per-branch `Run` tests; table-driven route tests; injection guards in SetupFn; TestMain seeding | High | [httptest](https://pkg.go.dev/net/http/httptest) · [Table-driven tests](https://go.dev/doc/table-driven-tests) · [goleak](https://pkg.go.dev/go.uber.org/goleak) |
| **Failure handling** | 4xx-vs-500 classification per handler; `verifySchema` fail-fast; `http.ErrServerClosed` as normal exit; Shutdown deadline; tx rollback pattern; panic-in-goroutine knowledge | High | [Errors are values](https://go.dev/blog/errors-are-values) · [Go 1.13 errors](https://go.dev/blog/go1.13-errors) · [Defer, panic, recover](https://go.dev/blog/defer-panic-and-recover) |
| **Scalability** | Runner pattern (concurrent run, reverse-order close); bounded-concurrency & backpressure patterns (studied, not yet applied); graceful drain; k8s shutdown contract | Medium | [Pipelines](https://go.dev/blog/pipelines) · [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) · [Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown) · [Proper HTTP shutdown](https://dev.to/mokiat/proper-http-shutdown-in-go-3fji) |
| **OO & patterns** | Strategy (`PermissionScope` ×5, read in operations-execution); Adapter (`FnRunner`); DI by constructor + SetupFn injection; decorator (two `mux.MiddlewareFunc`s = a decorator chain); registry-as-data; implicit interfaces; composition over inheritance | Medium-high | [SOLID Go Design (Cheney)](https://dave.cheney.net/2016/08/20/solid-go-design) · [Design patterns catalog](https://refactoring.guru/design-patterns) · [Effective Go](https://go.dev/doc/effective_go) |
| **API design** | `/api/v1` versioning; wire models as the shared contract; `WriteBadRequest`/`WriteInternalServerError` shape; X-Request-ID correlation | Medium | [Google API Design Guide](https://cloud.google.com/apis/design) · [RFC 9457 problem+json](https://www.rfc-editor.org/rfc/rfc9457) · [Stripe idempotency](https://stripe.com/blog/idempotency) · [Idempotency (MDN)](https://developer.mozilla.org/en-US/docs/Glossary/Idempotent) |

---

## 1. Concurrency (goroutines, channels, WaitGroup)

- [Pipelines and cancellation](https://go.dev/blog/pipelines) — canonical fan-out/fan-in; the worker-pool discussion
- [Go Concurrency Patterns: Context](https://go.dev/blog/context) — the Go team's context rationale; exactly R1
- [Share Memory By Communicating](https://go.dev/blog/codelab-share) — channel-vs-mutex philosophy
- [The Go Memory Model](https://go.dev/ref/mem) — happens-before; the quiescence argument for WaitGroup
- [errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup) — the production composite hand-built in `runner.Run`

## 2. Debugging & failure handling

- [Data Race Detector](https://go.dev/blog/race-detector) — `go test -race`, the highest-value tool
- [goleak](https://pkg.go.dev/go.uber.org/goleak) — goroutine-leak detection in tests
- [Defer, Panic, and Recover](https://go.dev/blog/defer-panic-and-recover) — defer evaluation rules; nil-receiver lesson
- [Errors are values](https://go.dev/blog/errors-are-values) + [Go 1.13 errors](https://go.dev/blog/go1.13-errors) — `errors.Is/As`, wrapping
- [Loop variable semantics](https://go.dev/wiki/LoopVariableSemantics) — the Go 1.22 capture gotcha (planted-bug #1)

## 3. Slices & maps internals (top gap — skim)

- [Slices: usage and internals](https://go.dev/blog/slices-intro) + [Go Slices: usage and internals](https://go.dev/blog/slices) — `append` growth, aliasing, len/cap
- [Go maps in action](https://go.dev/blog/maps) — random iteration order; nil maps safe to read

## 4. Testing (the three layers)

- [net/http/httptest](https://pkg.go.dev/net/http/httptest) — Recorder vs NewServer layering
- [Table-driven tests](https://go.dev/doc/table-driven-tests) — official pattern behind the route tests
- [Practical Go (Dave Cheney)](https://dave.cheney.net/practical-go/presentations/qcon-china.html) — single best pre-interview read: interfaces, errors, API design

## 5. API design & HTTP

- [Google API Design Guide](https://cloud.google.com/apis/design) — pagination, versioning, errors, resource naming
- [RFC 9457: Problem Details](https://www.rfc-editor.org/rfc/rfc9457) — `{type, title, status, detail, instance}`
- [Stripe: Designing robust APIs with idempotency](https://stripe.com/blog/idempotency) — the production answer to "is POST /create OK?"
- [Idempotency (MDN)](https://developer.mozilla.org/en-US/docs/Glossary/Idempotent) — status-code semantics
- [How I write HTTP services in Go after 13 years (Mat Ryer)](https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/) — handlers as `http.Handler`, constructor functions, graceful shutdown — mirrors this repo's shape

## 6. Graceful shutdown & lifecycle (R1/R3)

- [net/http Server.Shutdown](https://pkg.go.dev/net/http#Server.Shutdown) — the official contract implemented here
- [Proper HTTP shutdown in Go](https://dev.to/mokiat/proper-http-shutdown-in-go-3fji) — Shutdown-before-Serve subtleties
- [os/signal](https://pkg.go.dev/os/signal) — `signal.Notify` semantics behind `waitForSignal`
- [Graceful shutdown: signals, draining, deadlines](https://bytesizego.com/blog/graceful-shutdown-for-go-http-servers-signals-draining-and-deadlines) — k8s drain contract

## 7. Patterns & style (bonus)

- [SOLID Go Design (Dave Cheney)](https://dave.cheney.net/2016/08/20/solid-go-design) — OO design principles in Go terms
- [Uber Go Style Guide](https://github.com/uber-go/guide) — what good Go looks like, codified
- [Effective Go](https://go.dev/doc/effective_go) — interfaces, embedding, composition
- [Design patterns catalog](https://refactoring.guru/design-patterns) — language-agnostic mapping (Strategy, Adapter, Decorator…)
- [Go Proverbs](https://go-proverbs.github.io) — quotable one-liners for any design discussion
