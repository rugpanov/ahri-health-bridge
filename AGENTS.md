# ahri-health-bridge — Agent Guide

Go HTTP service that receives Apple Health data from iOS Shortcuts and stores it in Neon (PostgreSQL).

## Quick Start

```bash
cp .env.example .env   # fill in API_KEY and DATABASE_URL
go run .               # starts on :8080
go test ./...
```

## Tech Stack

- Go 1.26+, Chi router, godotenv, pgx/v5 + pgxpool, golang-migrate (embedded SQL)

## Structure

```
handlers/      # HTTP layer
controllers/   # Parse JSON, call logger + store
gateways/      # logger.go (file writer), db.go (HealthStore, NoopStore, NeonStore)
utils/         # auth.go (X-API-Key middleware), errs.go (ErrBadRequest sentinel)
migrations/    # SQL files embedded at build time, run on startup
main.go
```

## API

```
POST /health/steps
X-API-Key: <key>
{"steps": 1000}
```

`200` success · `400` bad payload · `401` bad key · `500` DB failure

## Config

```
API_KEY=<secret>
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
PORT=8080          # default
LOG_FILE=payloads.json   # default
LINEAR_API_KEY=<key>
```

## Extending

To add a new parameter (e.g. sleep): add migration, `handlers/sleep.go`, `controllers/sleep.go`, `StoreSleep` in `gateways/db.go`, one route in `main.go`.

## Git operations

See the [top-level AGENTS.md](../../AGENTS.md) for guidance on git pull/push, including SSH key troubleshooting.

## Conventions

- British spelling · no global state · interfaces in consumer packages · `.env` never committed

## Linear

Project: [ahri-health-bridge](https://linear.app/highloadapp/project/ahri-health-bridge-5c80c52ee823) · Team: `HIG`

All issues HIG-45–HIG-58 Done (Phase 1: Go service scaffold; Phase 2: Neon storage).

API key: `LINEAR_API_KEY` in `.env`, no `Bearer` prefix.

## Design Docs

- `docs/superpowers/specs/2026-05-18-ahri-health-bridge-design.md`
- `docs/superpowers/specs/2026-05-19-steps-neon-storage-design.md`
