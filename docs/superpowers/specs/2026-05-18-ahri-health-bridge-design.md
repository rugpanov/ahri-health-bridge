# ahri-health-bridge Design

**Date:** 2026-05-18  
**Status:** Approved

## Overview

A Go HTTP service that receives Apple Health data from iOS Shortcuts and stores it in Neon (PostgreSQL). Phase 1 is exploration-only: capture any raw payload, log it to stdout and a local file, so the actual schema can be understood before committing to a data model. Steps is the first health parameter; the architecture is designed to extend cleanly to sleep, heart rate, and others.

## Architecture

**Stack:** Go 1.22+, Chi router, godotenv  
**Deployment:** Local first, Fly.io-ready later

### Project Structure

```
ahri-health-bridge/
├── main.go              # entry point: loads config, wires router, starts server
├── handlers/
│   └── steps.go         # HTTP layer — reads body, calls controller
├── controllers/
│   └── steps.go         # business logic — validates, routes to gateways
├── gateways/
│   ├── logger.go        # stdout + payloads.json file writer
│   └── db.go            # Neon/PostgreSQL stub (interface only, not wired)
├── utils/
│   └── auth.go          # X-API-Key middleware
├── .env.example
├── go.mod
└── README.md
```

Adding a new health parameter means: one new file in each of `handlers/`, `controllers/`, and one new route in `main.go`. Nothing else changes.

## API

### POST /health/steps

Receives raw Apple Health steps data from an iOS Shortcut.

**Request:**
```
POST /health/steps
X-API-Key: <key>
Content-Type: (any — not enforced in phase 1)
Body: (any payload — raw capture, no schema enforced)
```

**Response:**
```json
{"status": "received"}
```

**Error responses:**
- `401 Unauthorized` — missing or invalid `X-API-Key`

## Data Flow

1. Auth middleware validates `X-API-Key` header.
2. `handlers/steps.go` reads the raw body bytes, calls the steps controller.
3. `controllers/steps.go` calls the logger gateway.
4. `gateways/logger.go` prints to stdout (timestamped, labelled `steps`) and appends a JSON line to `payloads.json`.
5. Handler returns `200 {"status": "received"}`.

`gateways/db.go` defines the storage interface but is a no-op stub in phase 1. It is wired up once the schema is understood from captured payloads.

## Config

Loaded from `.env` at startup via `godotenv`:

```
PORT=8080
API_KEY=<your-key>
LOG_FILE=payloads.json
```

## Error Handling

| Scenario | Behaviour |
|---|---|
| Missing/invalid `API_KEY` env var at startup | Server refuses to start (fail fast) |
| Wrong `X-API-Key` header | `401 Unauthorized` |
| Empty request body | Logged as-is, `200 OK` returned |
| File write failure | Error logged to stdout, `200 OK` still returned |

Empty bodies and file write failures are handled permissively during exploration — the priority is capturing data from Shortcuts without breaking the flow.

## Testing

No tests in phase 1. The service is exploratory. Tests are added once the schema is known and real controller logic exists.

## Extension Plan

When a new health parameter is ready (e.g. sleep, heart rate):
1. Add `handlers/<param>.go`
2. Add `controllers/<param>.go`
3. Add one route line in `main.go`
4. Implement the storage gateway in `gateways/db.go` for that param

Auth, logging, and config require no changes.
