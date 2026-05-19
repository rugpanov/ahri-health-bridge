# ahri-health-bridge — Agent Guide

Go HTTP service that receives Apple Health data from iOS Shortcuts, validates it, and stores it in Neon (PostgreSQL). Parses `{"steps": N}` from the request body; returns `400` for validation errors, `500` for DB failures.

## Quick Start

```bash
cp .env.example .env   # fill in API_KEY and DATABASE_URL
go run .               # starts on :8080
go test ./...          # run all tests
go build .             # build binary
```

## Tech Stack

- Go 1.26+, module `github.com/rugpanov/ahri-health-bridge`
- Router: `github.com/go-chi/chi/v5`
- Config: `github.com/joho/godotenv` (`.env` file)
- DB driver: `github.com/jackc/pgx/v5` + pgxpool
- Migrations: `github.com/golang-migrate/migrate/v4` with embedded SQL (`embed.FS`)

## Project Structure

```
handlers/          # HTTP layer — read body, call controller, return JSON
controllers/       # Business logic — parse JSON, call logger + store
gateways/          # External integrations
  logger.go        # stdout + payloads.json file writer
  db.go            # HealthStore interface, NoopStore (test double), NeonStore
utils/             # Shared utilities
  auth.go          # X-API-Key header validation middleware
  errs.go          # ErrBadRequest sentinel (handlers use this for 400 vs 500)
migrations/        # Embedded SQL migration files (run on startup)
main.go            # Entry point: load config, run migrations, wire dependencies, start server
```

## API

```
POST /health/steps
Headers: X-API-Key: <key>
Body:    {"steps": 1000}
Response: {"status":"received"}
```

| Code | Condition |
|---|---|
| `200` | Steps stored successfully |
| `400` | Missing or invalid `steps` field in JSON |
| `401` | Missing or wrong `X-API-Key` |
| `500` | Database write failure |

## Config (`.env`)

```
PORT=8080
API_KEY=<secret>
LOG_FILE=payloads.json
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
LINEAR_API_KEY=<linear-api-key>
```

Server refuses to start if `API_KEY` or `DATABASE_URL` is not set.

## Extending to New Health Parameters

To add a new parameter (e.g. sleep):
1. Add `migrations/000002_create_sleep.up.sql` + `.down.sql`
2. Add `handlers/sleep.go` with a `SleepHandler`
3. Add `controllers/sleep.go` with a `SleepController`
4. Add `StoreSleep` to `gateways/db.go`
5. Wire one route in `main.go`: `r.Post("/health/sleep", sleepHandler.ServeHTTP)`

Auth, logging, migrations, and config need no changes.

## Conventions

- British spelling in docs and comments
- No global state — all dependencies injected via constructors
- Interfaces defined in consumer packages (dependency inversion)
- `.env` is never committed; `payloads.json` is excluded by `.gitignore`

## Linear Project

Project: [ahri-health-bridge](https://linear.app/highloadapp/project/ahri-health-bridge-5c80c52ee823)
Team: Highloadapp (`highloadapp`)

**Phase 1** — raw payload capture

| Issue | Title | Status |
|---|---|---|
| HIG-45 | Task 1: Initialise Go module and install dependencies | Done |
| HIG-46 | Task 2: Auth middleware | Done |
| HIG-47 | Task 3: Logger gateway | Done |
| HIG-48 | Task 4: DB gateway stub | Done |
| HIG-49 | Task 5: Steps controller | Done |
| HIG-50 | Task 6: Steps handler | Done |
| HIG-51 | Task 7: Wire up main.go and smoke test | Done |

**Phase 2** — Neon storage

| Issue | Title | Status |
|---|---|---|
| HIG-52 | Task 1: Install pgx/v5 and golang-migrate deps | Done |
| HIG-53 | Task 2: Create steps migration files | Done |
| HIG-54 | Task 3: Add ErrBadRequest sentinel to utils | Done |
| HIG-55 | Task 4: Update gateways/db.go — NeonStore + interface | Done |
| HIG-58 | Task 5: Update controllers/steps.go — JSON parsing + context | Done |
| HIG-56 | Task 6: Update handlers/steps.go — context + 400 vs 500 | Done |
| HIG-57 | Task 7: Update main.go + .env.example + smoke test | Done |

API key: set `LINEAR_API_KEY` in `.env`. Auth header uses the key directly (no `Bearer` prefix).

## Design Docs

- Phase 1 spec: `docs/superpowers/specs/2026-05-18-ahri-health-bridge-design.md`
- Phase 1 plan: `docs/superpowers/plans/2026-05-18-ahri-health-bridge.md`
- Phase 2 spec: `docs/superpowers/specs/2026-05-19-steps-neon-storage-design.md`
- Phase 2 plan: `docs/superpowers/plans/2026-05-19-steps-neon-storage.md`
