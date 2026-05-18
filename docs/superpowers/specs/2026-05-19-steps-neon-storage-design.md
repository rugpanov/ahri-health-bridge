# ahri-health-bridge Phase 2: Steps Neon Storage Design

**Date:** 2026-05-19  
**Status:** Approved

## Overview

Phase 1 captured raw Apple Health payloads and logged them locally. The payload schema is now known: `{"steps": 0}`. Phase 2 parses this payload and stores step counts in a Neon (PostgreSQL) database. DB write failures return `500` so iOS Shortcuts can surface them.

## Architecture

Only three existing files change; two new migration files are added.

### Files Changed

| File | Change |
|---|---|
| `gateways/db.go` | Add `NeonStore` (pgxpool); keep `NoopStore` as test double |
| `controllers/steps.go` | Parse `{"steps": N}`; call logger then `StoreSteps` |
| `main.go` | Add `DATABASE_URL` config, init pgxpool, run migrations, wire `NeonStore` |
| `.env.example` | Add `DATABASE_URL` |

### Files Added

| File | Purpose |
|---|---|
| `migrations/001_create_steps.sql` | UP migration — create `steps` table |
| `migrations/001_create_steps.down.sql` | DOWN migration — drop `steps` table |

### Tech Stack Additions

- `github.com/jackc/pgx/v5` + `pgxpool` — Neon-recommended PostgreSQL driver with connection pooling
- `github.com/golang-migrate/migrate/v4` — migration runner
- `migrate/v4/database/pgx/v5` — pgx driver adapter for golang-migrate
- `migrate/v4/source/iofs` — embed.FS source adapter

## Data Model

```sql
CREATE TABLE IF NOT EXISTS steps (
    id          BIGSERIAL PRIMARY KEY,
    steps       INTEGER NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

`received_at` defaults to `NOW()` server-side — the controller does not pass a timestamp.

## Migrations

SQL files are embedded into the binary using Go's `embed.FS`. Migrations run on startup in `main.go` before the server accepts connections. Migration failure causes a fatal exit.

## Data Flow

```
POST /health/steps
  → auth middleware (X-API-Key)
  → handler reads raw body
  → controller parses {"steps": N} → extracts int
      ↳ malformed JSON or missing field → 400
  → logger.Log("steps", body)           [always, raw bytes]
  → StoreSteps(ctx, stepsCount)         [parsed int]
      ↳ DB failure → log to stdout + return 500
  → 200 {"status":"received"}
```

## Error Handling

| Scenario | Behaviour |
|---|---|
| Missing/invalid `DATABASE_URL` at startup | Fatal exit |
| Migration failure at startup | Fatal exit |
| Malformed JSON body | `400 Bad Request` |
| Missing `steps` field in JSON | `400 Bad Request` |
| DB write failure | Stdout log + `500 Internal Server Error` |
| Logger file write failure | Stdout log only, DB write still attempted |

## Config

Add to `.env` and `.env.example`:

```
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
```

## Interface Contract

`HealthStore` interface in `gateways/db.go` changes to accept a typed value:

```go
type HealthStore interface {
    StoreSteps(ctx context.Context, steps int) error
}
```

The controller parses `{"steps": N}` once, calls `logger.Log` with the raw body, then calls `StoreSteps(ctx, stepsCount)` with the parsed integer. The gateway receives a clean value — no JSON parsing in the storage layer.

`NoopStore` is updated to match the new signature and remains in `gateways/db.go` for use in controller tests.
