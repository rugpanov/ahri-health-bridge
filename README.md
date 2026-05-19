# ahri-health-bridge

Go HTTP service that receives Apple Health data from iOS Shortcuts and stores it in Neon (PostgreSQL).

## Quick Start

```bash
cp .env.example .env   # fill in API_KEY and DATABASE_URL
go run .               # starts on :8080
go test ./...          # run all tests
go build .             # build binary
```

## API

```
POST /health/steps
Headers: X-API-Key: <key>
Body:    {"steps": 1000}
```

| Response | Condition |
|---|---|
| `200 {"status":"received"}` | Steps stored successfully |
| `400` | Missing or invalid `steps` field |
| `401` | Missing or wrong `X-API-Key` |
| `500` | Database write failure |

## Config (`.env`)

```
PORT=8080
API_KEY=<secret>
LOG_FILE=payloads.json
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
```

## Tech Stack

- Go 1.26+, Chi router, godotenv
- `github.com/jackc/pgx/v5` + pgxpool — Neon-recommended PostgreSQL driver
- `github.com/golang-migrate/migrate/v4` — embedded SQL migrations

## Extending

To add a new health parameter (e.g. sleep):
1. Add `handlers/sleep.go`, `controllers/sleep.go`
2. Add `StoreSleep` to `gateways/db.go`
3. Add `migrations/000002_create_sleep.up.sql`
4. Wire one route in `main.go`
