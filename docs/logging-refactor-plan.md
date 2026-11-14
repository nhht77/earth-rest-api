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

## Phases (each phase = one green `go build && go test ./...`)

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

### Phase 3 — Inject the logger (remove the global)
- [ ] `Config`/`Database`/handlers: add `Log logging.Logger` field.
- [ ] `main.go`: construct once, inject everywhere (composition-root style:
      only `main.go` knows `NewLogger`).
- [ ] `pkg/mhttp`: take `log logging.Logger` parameter (it's a helper package
      — constructor `mhttp.New(log)`).
- [ ] Grep proves zero package-global logger access.
- [ ] `main_test.go`: inject a capturing test logger
      (add `loggingtest` variant: logger writing to a bytes buffer with
      `GetEntries()`-style access).

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

## Out of scope (for later plans)
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