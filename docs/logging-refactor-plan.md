# Logging refactor plan — earth-rest-api

> Goal: turn the raw-logrus global (`Log`) into an injected, configured,
> structured logger — the `pkg/logging` pattern from operations-execution,
> implemented locally (no Nokia internal imports; this is a public repo).
> This is study exercise for weakness #5 (logging) and #6 (DI).

## Current state (evidence)

- `server/main.go`: package-global `Log = logrus.New()`, `TextFormatter`,
  output to **stdout**, fixed config in `init_logger()`.
- 24 `Log.*` call sites across 8 files: `config.go`, `database*.go`,
  `http_server.go`, `pkg/mhttp/mhttp.go`, `main_test.go`.
- No levels config, no `logger` (source) field, no stack traces on Error,
  no request correlation, no metrics.

## Target state

- A local `logging` package (`server/pkg/logging`) with a small `Logger`
  interface — same shape as Neo v1: `Debug/Info/Warn/Error/Fatal` (+f/ln),
  `With(key, value) Logger`.
- Config from env: `LOGGING_LEVEL` (debug|info|error), `LOGGING_FORMAT`
  (json|txt). Default: info, json.
- Output to **stderr**, JSON with ISO8601 timestamps.
- Every entry carries `logger: file.go:NN` (call-site via `runtime.Caller`).
- `Error*` auto-attaches `stack_trace`.
- Handlers/DB access a logger via injection (struct field / `context`), not
  a global.

## Phase order — decision record (2026-11-15)

Preference: **modularize `database/` and `http_server.go` before polishing
minor logging things.** Revised order:

- [x] ~~Phase 3 — Inject the logger (remove the global)~~ done 15.11.2026
- [x] **Phase 3.5 — Modularize (PREFERRED NEXT)** done 18.11.2026
      - `database/` package: `database.go` + `database_{city,country,continent}.go`
        + `Config` DB wiring → own package with constructor (`NewDatabase(log)`)
      - `endpoints/` (or `http/`) package: `http_server.go` + `http_handlers_*.go`
        + `handlers.go` → own package; `RunHTTP(h, log)` becomes the package's
        constructor/entrypoint
      - `app/` package (new file): composition root — extract
        `init_resource`/`init_framework`/`RunHTTP` glue from `main.go`;
        `main.go` shrinks to the ~20-line script (build logger, wire signal,
        app.Setup, app.Run)
      - `pkg/logging` context helpers already exist (NewContext/FromContext) —
        middleware can move with http_server.go later (Phase 4 folds into this)
- [x] **Phase 4 — request correlation** (folds into 3.5's `endpoints/` package:
      request-ID middleware, FromContext in handlers, monitor line with
      duration+status)
- [x] **Phase 5 — error classification + stack trace polish** (DB errors → 500
      in handlers; `exit` var already in place for Fatal)

Rationale: modularizing now (while components are freshly injected) is the
mechanical, high-value step; Phase 4/5 polish is better done once files live
in their final packages so the work isn't moved twice.

## Phase order (original)

### Phase 1 — Create the logging package (additive, no callers changed)
- [ ] `server/pkg/logging/logging.go`: `Logger` interface + `logger` struct
      wrapping `*logrus.Entry` (copy the interface from
      `pkg/logging/logging.go` in operations-execution, trimmed to what
      earth-rest-api uses: no `Print` for now).
- [ ] `NewLogger()` reading env vars, defaulting info/json, writing stderr.
- [ ] `sourced(depth)` helper → `logger` field.
- [ ] Test: `logging_test.go` asserts JSON output shape via a buffer
      (`logrus.Logger.SetOutput(&bytes.Buffer{})`).

### Phase 2 — Swap the global, keep call sites compiling
- [ ] Keep `var Log logging.Logger = logging.NewLogger()` in `main.go` as the
      *only* remaining global (temporary).
- [ ] Replace raw logrus calls with the interface (type-compatible: existing
      `Log.Warnf` etc. keep working).
- [ ] Delete `init_logger()`; main calls `logging.NewLogger()`.
- [ ] Verify output changed: run server, see JSON on stderr with
      `logger: file.go:NN` fields.

### Phase 3 — Inject the logger (remove the global) ✅ done 15.11.2026
- [x] `Config`/`Database`/handlers: add `Log logging.Logger` field.
- [x] `main.go`: constructs once, inject everywhere (composition-root style:
      only `main.go` knows `NewLogger`).
- [x] `pkg/mhttp`: verified pure (no logging) — response helpers stay
      parameter-only by design.
- [x] Handlers converted to `*Handlers` struct methods (15 handlers across
      city/country/continent files); `RunHTTP(h *Handlers, log)` and
      `MonitorMiddleware(log)` factory pattern.
- [x] Database built with field init: `DB = &Database{Log: Log}` (nil-field
      panic encountered and fixed — lesson banked).
- Remaining: payoff test (construct `Database` with test logger and assert
  on captured entries) — can land with Phase 4's loggingtest helper.

### Phase 4 — Request correlation (middleware + context)
- [ ] Add a `RequestLogger` middleware: create `logger.With(request_id,
      method, path)` per request, store in `r.Context()`
      (`logging.LoggerFromContext`), retrieve in handlers.
- [ ] `MonitorHandle` now logs one line with duration + status code
      (needs a small response-writer wrapper capturing status).
- [ ] Handlers stop logging "I got a request" — middleware owns that line.

### Phase 5 — Error-level quality
- [ ] `Error*` methods attach `stack_trace` (port from operations-execution
      `logging.go`).
- [ ] Classify handler errors: DB failure → 500 + `Error` log at the handler
      boundary; bad input → 400, no stack. This also preps weakness #1
      (error handling) work.

## Credentials concern (local → containerized → Helm)

Context: earth-rest-api reads DB credentials from env vars (`APP_USERNAME`,
`APP_PASSWORD`, `APP_DATABASE`) sourced from a plain `env` file at the repo
root (see README). The repo has **no Dockerfile / Helm chart yet**, so the
container stages below are requirements to keep in mind when that work
happens — the point is that the `env` file is a *local-only* mechanism and
nothing downstream of it may depend on a file existing.

### Local run (today)
- `env` file with `export` lines, sourced before `./server`:
  `source env && ./server`
- The file must stay **gitignored**; DB user should be least-privilege
  (app schema only)

### When the application is containerized (future)
- **No env file exists inside a container** — env vars must be injected at
  `docker run --env-file env` time, never `-e APP_PASSWORD=...` (shell
  history / `ps` leak)
- Add `env` to `.dockerignore` so a careless `COPY . .` never bakes
  credentials into an image layer (immutable artifact, leaks forever)
- Never log the config struct wholesale (`Log.Info(AppConfig)` is the
  classic leak) — this intersects with the logging refactor: after
  Phase 5, sensitive values should be wrapped with a privacy marker
- Keep `Config` parsing (envconfig-style) as the single ingress for
  credentials so the container path has exactly one place to audit

### When packaged into a Helm chart (future)
- **No real credentials in `values.yaml`** — values files get committed;
  a default password there *is* a leak
- The chart should *consume* a pre-existing Kubernetes Secret rather than
  create one from values:
  `envFrom: [{secretRef: {name: earth-db-credentials}}]`
- Options for the secret itself, increasing sophistication:
  1. Pre-provisioned Secret out-of-band (documented name in chart README)
  2. `helm-secrets` + SOPS-encrypted values
  3. External Secrets Operator / Vault
- Never `helm install --set dbPassword=...` — lands in shell history and
  the rendered release manifest
- Non-sensitive config (port, feature flags) → ConfigMap; credentials →
  Secret (base64 is not encryption)

### One-liner for the interview
"Same three knobs (username/password/database), different injection per
runtime: env file locally, env vars from the container runtime, and a
pre-existing Secret referenced by the Helm chart — with values files
carrying only non-secret defaults."

## Architecture notes: repo organization (pkg/svc monorepo pattern)

Observation while studying: the AANM Go monorepo (`~/dev/go`) splits code as
`pkg/` (shared libraries: logging, tracing, kafkatable, mariadb, runner —
no `main()`, not deployable) and `svc/` (deployables: operations-execution,
operations-feedback — each with own Dockerfile, Jenkinsfile, helm/).
Jenkins builds each `svc/*` independently.

Key distinction learned:

| Axis | Question | Example |
|---|---|---|
| Deployable boundary | Should this *run* as a separate service? | operations-execution vs operations-feedback — justified by ownership/scale/isolation |
| Repo organization | How to *arrange code* for services? | `pkg/` + `svc/` in one repo |

The monorepo does NOT imply splitting the domain — cities/continents stay in
one service regardless. Split deployables on team/scale boundaries only.

### AANM monorepo mechanics
- **One root `go.mod`** — all services share dependency versions; a `pkg/`
  change is instantly visible everywhere (plus: one CI pipeline, atomic
  cross-service PRs; minus: lockstep upgrades, CI blast radius)
- **Cross-service contracts as code**: generated clients under
  `pkg/operations/api/*`, `contract_test/` consumer suites
- **`svc/<name>/` self-contained**: `cmd/`, `app/`, `endpoints/`, `helm/`,
  `Dockerfile`, `Jenkinsfile`

### Earth-rest-api mapping (right-sized)
Today it has one deployable, so full pkg/svc is overkill. The honest layout:

```
earth-rest-api/
├── pkg/       ← shared: logging (new), mhttp, msql, muuid (exist)
└── server/    ← the one deployable (rename to svc/ only if a 2nd appears)
```

Teaches the layout without pretending the domain justifies multiple services.

### Decision record (why NOT a microservice split of the domain)
- cities → countries → continents is a real relational chain (`JOIN`s today,
  network hops + cross-service transactions if split)
- three deployables for a tiny domain = deployment/observability overhead
  without a scaling need — operations-execution's own README documents the
  distributed-state pain (Kafka + lease manager) that splitting creates
- revisit only if: separate team ownership, independent scaling needs, or
  isolated failure domains emerge

### LATE REFACTOR note (2026-11-15): shared framework vs per-service ownership
Discussion outcome (keep for a future split — not part of current phases):

- If the domain is ever split into city/country/continent services, do NOT
  share `database/` across services (one shared pkg/database = shared schema
  = lockstep deploys = distributed monolith). Each service owns its own
  database/queries + config keys + routes.
- What IS shared lives in `pkg/`: logging, mhttp (router bootstrap,
  Request/Monitor middleware, ListenAndServe), msql (connect/ping/migrate
  runner), muuid — mechanism only, no domain.
- AANM monorepo evidence: `svc/*/cmd/main.go` md5s all differ — the "common
  main.go" is a copy-paste template convention, NOT inherited code. Shared
  extension points live in `pkg/runner` (`app.Context{Closers, Runners:
  []runner.Runnable{}}`); each svc's `app.Setup` fills them in.
- Rule to remember: "share pkg, copy the template, per-service ownership of
  app/cmd/database/endpoints." pkg exports extension points; services
  implement use-cases; main is per-service glue, never shared.
- Mechanism/use-case split inside current code when extracting:
  `MonitorMiddleware`/`ListenAndServe` → pkg/mhttp-shaped mechanism;
  the `router.HandleFunc(...)` table → use-case.
- ALSO modernize style to match AANM / operations-execution:
  - HTTP: `go-chi/chi/v5` (ops-exec router; gorilla/mux already legacy there)
    + shared `pkg/httpmiddleware`; consider OpenAPI-generated handler stubs
    (ops-exec endpoints are generated from openapi.yaml)
  - Database: interface-first — define `Datastorer`-style interface with
    `ctx context.Context` on every method, typed returns; concrete
    (mariadb/oracle-style) implementation behind it; handlers depend on the
    interface → testable with mocks
  - General: ctx propagation everywhere, runner-style lifecycle
    (Closers/Runners in app.Context), pkg/ exports extension points

## Out of scope (for later plans)
- **IT (integration tests)**: a dedicated `server/integration_test/` package
  with `//go:build integration` tag (operations-execution pattern) covering
  the DB-backed handlers and `database/` query methods against `earth_test`.
  `main_test.go`'s DB-backed `TestMain` setup is the seed for it — move it
  there when doing this. Unit suites (`config`, `handlers`, `app`) stay
  colocated and tag-free.
- Prometheus metrics hook (needs a metrics server in this app).
- Audit logging (no audit events in this app).
- v2 context-first API (worth a follow-up plan — mirrors Neo v2 with
  `ctx` as first parameter everywhere).

## Definition of done
- `grep -rn "logrus" --include="*.go" server | grep -v pkg/logging` → empty
  (only the logging package imports logrus).
- All logs are JSON on stderr with `logger` field; request logs carry
  `request_id` + duration.
- `main_test.go` demonstrates asserting on captured log entries.
- Every phase documented in the progress log of the fast-track doc.

# own note
take note about error handling
refactor main
refactor app
refactor database
refactor http

dockerize
helm chart builder

# live interview prep with golang chat gpt
+ problem
+ solution

+ stack/queue (can be cache) 
+ goroutine in different scenario (parall, async)
+ golang Context -> stop function with graceful shutdown