# ahri-health-bridge — Agent Guide

Go HTTP service that receives Apple Health data from iOS Shortcuts and logs it locally. Phase 1: capture any payload from steps, log to stdout + file. Designed to extend to sleep, heart rate, etc.

## Quick Start

```bash
cp .env.example .env   # fill in API_KEY
go run .               # starts on :8080
go test ./...          # run all tests
go build .             # build binary
```

## Tech Stack

- Go 1.26+, module `github.com/rugpanov/ahri-health-bridge`
- Router: `github.com/go-chi/chi/v5`
- Config: `github.com/joho/godotenv` (`.env` file)

## Project Structure

```
handlers/          # HTTP layer — read body, call controller, return JSON
controllers/       # Business logic — validate, call gateways
gateways/          # External integrations
  logger.go        # stdout + payloads.json file writer
  db.go            # HealthStore interface + NoopStore stub (for future Neon/PostgreSQL)
utils/             # Shared middleware
  auth.go          # X-API-Key header validation
main.go            # Entry point: load config, wire dependencies, start server
```

## API

```
POST /health/steps
Headers: X-API-Key: <key>
Body: any payload (raw capture, no schema enforced in phase 1)
Response: {"status":"received"}
```

Returns `401` if `X-API-Key` is missing or wrong.

## Config (`.env`)

```
PORT=8080
API_KEY=<secret>
LOG_FILE=payloads.json
LINEAR_API_KEY=<linear-api-key>
```

Server refuses to start if `API_KEY` is not set.

## Extending to New Health Parameters

To add a new parameter (e.g. sleep):
1. Add `handlers/sleep.go` with a `SleepHandler`
2. Add `controllers/sleep.go` with a `SleepController`
3. Add one route in `main.go`: `r.Post("/health/sleep", sleepHandler.ServeHTTP)`
4. Implement `StoreSleep` in `gateways/db.go` when DB is ready

Auth, logging, and config need no changes.

## Conventions

- British spelling in docs and comments
- No global state — all dependencies injected via constructors
- Interfaces defined in consumer packages (dependency inversion)
- `.env` is never committed; `payloads.json` is excluded by `.gitignore`

## Linear Project

Project: [ahri-health-bridge](https://linear.app/highloadapp/project/ahri-health-bridge-5c80c52ee823)
Team: Highloadapp (`highloadapp`)

| Issue | Title | Status |
|---|---|---|
| HIG-45 | Task 1: Initialise Go module and install dependencies | Done |
| HIG-46 | Task 2: Auth middleware | Done |
| HIG-47 | Task 3: Logger gateway | Done |
| HIG-48 | Task 4: DB gateway stub | Done |
| HIG-49 | Task 5: Steps controller | Done |
| HIG-50 | Task 6: Steps handler | Done |
| HIG-51 | Task 7: Wire up main.go and smoke test | In Progress |

API key: set `LINEAR_API_KEY` in `.env`. Auth header uses the key directly (no `Bearer` prefix).

## Design Docs

- Spec: `docs/superpowers/specs/2026-05-18-ahri-health-bridge-design.md`
- Plan: `docs/superpowers/plans/2026-05-18-ahri-health-bridge.md`
