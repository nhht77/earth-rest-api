# Service evolution plan — earth-rest-api

Created: 18.11.2026. Supersedes nothing; complements `logging-refactor-plan.md`
(the logging/DI/modularization phases there are done — this plan covers the
next evolutionary steps).

Sources studied (facts, not speculation):

- `~/dev/go` AANM monorepo:
  - `svc/operations-execution/` — `cmd/main.go`, `app/app.go` (Context struct,
    `SetupFn` option list, `runner.Run(ctx, Runners...)`), `migrations/`
    (numbered `up`/`down` SQL), `integration_test/` + `contract_test/`
    (build-tagged), `httpmiddleware/`
  - `svc/operations-feedback/` — go-swagger generated `restapi/` + hand-written
    `controller/`, `db/{mariadb,mock}`, own `cmd/producer`, `cmd/storage` binaries
  - root `pkg/` — shared libraries incl. `runner` (Runnable interface),
    `logging`, `mariadb`, `oracle`
  - `Dockerfile.migrations` — schema migrations run by a dedicated
    `service-db-migrate` image, NOT by the application process
  - Makefile: `INT_TEST_FLAGS = -tags=integration -run='^TestIntegration'`,
    `DB_PREFIX=int` — isolated DBs per integration run

## 1. Runtime structure (main.go / app.go, AANM style)

Goal: `cmd`-level main that only orchestrates lifecycle; `app` package that
owns wiring; context as the cancellation spine; graceful shutdown.

Target shape:

```
server/
├── main.go            ← lifecycle script only (AANM cmd/main.go pattern):
│                        ctx + signal handler → app.New → Setup(ctx) → Run(ctx)
└── app/
    ├── app.go         ← App struct = AANM's Context: Closers []func(),
    │                    Runners []Runnable; Setup = ordered list of setup fns
    └── runner.go      ← small local runner (pkg/runner equivalent, ~40 lines)
```

Phases (each = one green build+test):

- [ ] **R1 — context spine + graceful shutdown**
  - `main.go`: `ctx, cancel := context.WithCancel(...)`, `go waitForSignal(ctx, cancel, log)`
    (SIGINT/SIGTERM → cancel — copy the AANM shape)
  - `app.Setup(ctx) error` / `app.Run(ctx) error` signatures; `Setup` stores ctx
  - `Run`: `http.Server` in a goroutine; `select { case <-ctx.Done(): srv.Shutdown(10s timeout) }`;
    `http.ErrServerClosed` is a *normal* exit, not an error
  - acceptance: kill -TERM → in-flight request completes, server exits 0
- [ ] **R2 — SetupFn option pattern**
  - `Setup` becomes an ordered list: `setupLogger → setupConfig → setupDB → setupHandlers → setupHTTPServer`
  - each `SetupFn` is idempotent (skip if field already set) — enables test
    harnesses to inject fakes between steps (exactly why ops-exec does it)
- [ ] **R3 — runners**
  - local `Runnable` interface (`Run() error; Close() error`); the HTTP server,
    and later any background worker, become runners
  - `app.Run = runner.Run(ctx, runners...)` — concurrent run, error aggregation
  - `Close()` of App collapses into runner Closers (deferred in Run, like AANM)
- [ ] **R4 — config: flags → env (envconfig-style)**
  - AANM uses `envconfig` env-vars (`PORT`, `DB_TYPE`), not CLI flags — k8s-native
  - keep the `NewFlagSet` path as a transitional fallback; add env override
  - decision to make: flags for local dev + env for containerized (documented in
    creds-concern section of logging-refactor-plan)

Interview story: *"I studied operations-execution's app.go — Setup as ordered
SetupFn list, Runners, closers — and ported the parts a small service earns:
context spine, graceful shutdown, option-style setup. I skipped the parts it
earns for its scale (k8s manager, lease manager, Kafka streams)."*

## 2. IT (integration test) plan — HTTP + database

Convention (copied from AANM Makefile):

- Location: `server/integration_test/` package, `//go:build integration` tag
- Run: `go test -tags=integration -run='^TestIntegration' ./...` from `server/`
- Isolated schema per run: connect to a dedicated test database
  (`earth_it_<runid>`-style or `DB_PREFIX`) so runs never collide
- `TestMain` there owns: create/patch schema, run migrations, tear down

Planned suites:

- [ ] **IT1 — database suite**: the real queries (CitiesByOptions with
  joins/triggers, CreateCity tx path, soft delete + trigger effects) against
  live Postgres. Seed: `main_test.go`'s current TestMain is the seed — move it
- [ ] **IT2 — HTTP suite**: full stack — router + middleware + handlers + DB;
  assert status codes (incl. the Phase-5 400/500 classification), headers
  (`X-Request-ID`), and JSON shapes end-to-end
- [ ] **IT3 — migration suite**: fresh empty DB → run migrations → schema
  matches what the queries expect (protects the schema-responsibility move in
  section 4)
- [ ] make targets: `make test-it` (spins up nothing new — uses
  earth-rest-api-db), CI: run with its own DB prefix

Unit tests stay colocated + tag-free; IT lives behind the tag — the AANM rule.

## 3. Splitting city / country / continent into own services + databases

Decision record (logging-refactor-plan): NO split today — revisit when a
forcing function appears. This section is the *plan for when we do it*, so the
seam is already chosen. Sequenced to be reversible at every step:

- [ ] **S1 — domain-scoped stores (in-repo, one DB)**
  - split `Database` into `CityStore`, `CountryStore`, `ContinentStore` sharing
    one connection; each gets its own `Datastorer`-style interface (this is the
    late-refactor item, realized per-domain)
  - acceptance: handlers depend only on their domain's interface; DB timing
    lines stay identical
- [ ] **S2 — schema-per-domain in one Postgres instance**
  - `city.`, `country.`, `continent.` schemas (Postgres schema ≠ database) with
    separate migrations dirs; one instance, three namespaces
  - the relational coupling becomes explicit: `city` keeps
    `country_uuid`/`continent_uuid` FKs — document what crosses the boundary
- [ ] **S3 — cross-boundary reads become explicit**
  - `ToCities` joins countries+continents: after S2, the join crosses schemas
    (still cheap, same instance); *plan the future choice*: embedded copies
    (denormalize at write) vs API reads. Default: embedded copies — reads stay
    local, writes publish updates (Kafka-outbox pattern, ops-exec's stream
    machinery is the reference)
- [ ] **S4 — separate databases** (only if needed): one Postgres per domain;
  cross-schema joins are gone by design after S3
- [ ] **S5 — separate deployables** (`server/city-service/`, own cmd/main,
  Dockerfile, helm) — last, only when team/scale/isolation justifies it
  (the ops-exec/ops-feedback precedent: each svc has its own everything)

Invariant through all phases: unit tests stay green; the wire format
(`pkg/city`, `pkg/country`, `pkg/continent`) is the shared contract — already
extracted, which is exactly what makes S1 cheap.

## 4. Schema responsibility → out of the application

Today: `Database.Initialize` runs `01-create-table.sql` etc. on every start.
That conflates "open connection" with "own the schema" — and AANM explicitly
does NOT do this (migrations are a separate image + CI step).

- [x] **C1 — local dev: schema from the postgres image** (done 18.11.2026)
  - postgres image auto-executes `/docker-entrypoint-initdb.d/*.sql` on first
    init: mount our migrations there (one-time, on `earth-rest-api-db` recreate)
  - acceptance: `docker compose up`-style start produces a complete schema with
    zero application migration runs
  - implemented: `migrations/` at repo root (moved from `server/sql/` via
    `git mv`); `make db-init` recreates `earth-rest-api-db` (port 5433) with
    the mount; `-- condition:` markers are just comments to psql — on a fresh
    DB all `CREATE TRIGGER` statements run
- [x] **C2 — app stops owning migrations** (done 18.11.2026)
  - `Initialize` keeps: validate config, connect, ping
  - migration execution moves out: startup *verifies* schema presence
    (cheap SELECT) and fails fast with a clear message if missing — pointing at
    the migration step, not silently self-healing
  - implemented: `verifySchema()` counts city/country/continent in
    `information_schema.tables`; the migration loop, `msql.ReadStatementsFromFile`
    condition parsing, and `resultChange` are deleted from the app
- [ ] **C3 — migration runner for CI/prod**
  - numbered `NNNN_name.up.sql`/`.down.sql` files (rename current files), run
    by a dedicated step — locally `make migrate`, in CI a `service-db-migrate`
   -style image/container (AANM `Dockerfile.migrations` precedent)
  - acceptance: two binaries of different versions can run against a migrated
    schema without re-running DDL (idempotency by version table)
- [ ] **C4 — IT3 (from section 2) guards this**: fresh DB + migrations-only =
  schema correct

## Sequencing recommendation

R1 (context spine) → C1+C2 (schema out of app — small, unblocks honest
bootstrap) → IT1+IT2 (tests gain power once the app is honest about lifecycle)
→ R2+R3 → S1 → (pause; S2–S5 wait for the forcing function) → R4.

Rationale: R1 makes the lifecycle real before more moving parts arrive; C1/C2
shrink the app's blast radius early; the service split (S*) starts as pure
in-repo refactoring whose payoffs (interfaces, explicit boundaries) are useful
even if S5 never happens.
