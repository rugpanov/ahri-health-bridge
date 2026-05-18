# Steps Neon Storage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add PostgreSQL storage to the steps endpoint — parse `{"steps": N}` from the request body, insert the value into a Neon `steps` table, and return `400` for validation errors or `500` for DB failures.

**Architecture:** `pgx/v5` + `pgxpool` for the connection; `golang-migrate` with embedded SQL files for schema management. Migrations run on startup. The controller parses once and passes a typed `int` to the store. A shared `ErrBadRequest` sentinel in `utils` lets the handler distinguish `400` vs `500` errors without importing the controllers package.

**Tech Stack:** Go 1.26+, `github.com/jackc/pgx/v5`, `github.com/golang-migrate/migrate/v4`, existing chi + godotenv

---

## File Map

| File | Change |
|---|---|
| `go.mod` / `go.sum` | Add pgx/v5 + golang-migrate deps |
| `migrations/000001_create_steps.up.sql` | CREATE TABLE steps |
| `migrations/000001_create_steps.down.sql` | DROP TABLE steps |
| `utils/errs.go` | Add `ErrBadRequest` sentinel |
| `utils/errs_test.go` | Test sentinel wrapping |
| `gateways/db.go` | Update `HealthStore` interface; add `NeonStore`; update `NoopStore` |
| `gateways/db_test.go` | Test `NoopStore` with new signature |
| `controllers/steps.go` | Add `StepStore` interface; parse JSON; pass `ctx`; call store |
| `controllers/steps_test.go` | Full replacement — new mock store + updated tests |
| `handlers/steps.go` | Add `ctx` to interface; distinguish 400 vs 500 |
| `handlers/steps_test.go` | Add validation error test; update mock signature |
| `main.go` | Add DATABASE_URL; embed migrations; init NeonStore; run migrations |
| `.env.example` | Add `DATABASE_URL` |

---

## Task 1: Install dependencies

**Files:**
- Modify: `go.mod`, `go.sum`

- [ ] **Step 1: Install pgx/v5 and golang-migrate**

```bash
cd tools/ahri-health-bridge
go get github.com/jackc/pgx/v5
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/pgx/v5
go get github.com/golang-migrate/migrate/v4/source/iofs
```

Expected: all four packages appear in `go.mod`

- [ ] **Step 2: Verify build still passes**

```bash
go build .
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add pgx/v5 and golang-migrate dependencies"
```

---

## Task 2: Create migration files

**Files:**
- Create: `migrations/000001_create_steps.up.sql`
- Create: `migrations/000001_create_steps.down.sql`

- [ ] **Step 1: Create migrations directory and up migration**

```bash
mkdir -p migrations
```

Create `migrations/000001_create_steps.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS steps (
    id          BIGSERIAL PRIMARY KEY,
    steps       INTEGER NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 2: Create down migration**

Create `migrations/000001_create_steps.down.sql`:

```sql
DROP TABLE IF EXISTS steps;
```

- [ ] **Step 3: Commit**

```bash
git add migrations/
git commit -m "feat: add steps table migration"
```

---

## Task 3: Add ErrBadRequest sentinel to utils

**Files:**
- Create: `utils/errs.go`
- Create: `utils/errs_test.go`

- [ ] **Step 1: Write the failing test**

Create `utils/errs_test.go`:

```go
package utils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

func TestErrBadRequest_IsWrappable(t *testing.T) {
	wrapped := fmt.Errorf("%w: invalid JSON", utils.ErrBadRequest)

	if !errors.Is(wrapped, utils.ErrBadRequest) {
		t.Error("expected errors.Is to match ErrBadRequest through wrapping")
	}
}
```

- [ ] **Step 2: Run to confirm it fails**

```bash
go test ./utils/...
```

Expected: compilation error — `utils.ErrBadRequest` undefined

- [ ] **Step 3: Implement**

Create `utils/errs.go`:

```go
package utils

import "errors"

// ErrBadRequest signals that the request payload is invalid.
// Handlers should return HTTP 400 when this error is present.
var ErrBadRequest = errors.New("bad request")
```

- [ ] **Step 4: Run to confirm it passes**

```bash
go test ./utils/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/utils`

- [ ] **Step 5: Commit**

```bash
git add utils/errs.go utils/errs_test.go
git commit -m "feat: add ErrBadRequest sentinel to utils"
```

---

## Task 4: Update gateways/db.go — NeonStore + interface

**Files:**
- Modify: `gateways/db.go`
- Create: `gateways/db_test.go`

The `HealthStore` interface changes from `StoreSteps(ctx, body []byte)` to `StoreSteps(ctx, steps int)`. `NoopStore` is updated to match. `NeonStore` is added.

- [ ] **Step 1: Write the failing test**

Create `gateways/db_test.go`:

```go
package gateways_test

import (
	"context"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/gateways"
)

func TestNoopStore_StoreSteps(t *testing.T) {
	store := &gateways.NoopStore{}
	err := store.StoreSteps(context.Background(), 100)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestNoopStore_StoreSteps_Zero(t *testing.T) {
	store := &gateways.NoopStore{}
	err := store.StoreSteps(context.Background(), 0)
	if err != nil {
		t.Errorf("expected no error for zero steps, got: %v", err)
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

```bash
go test ./gateways/...
```

Expected: compilation error — `NoopStore.StoreSteps` has wrong signature

- [ ] **Step 3: Replace gateways/db.go**

Overwrite `gateways/db.go` with:

```go
package gateways

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthStore is the interface for persisting health data.
// Implement a real version when Neon storage is ready.
type HealthStore interface {
	StoreSteps(ctx context.Context, steps int) error
}

type NoopStore struct{}

func (n *NoopStore) StoreSteps(_ context.Context, _ int) error {
	return nil
}

type NeonStore struct {
	pool *pgxpool.Pool
}

func NewNeonStore(ctx context.Context, connString string) (*NeonStore, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &NeonStore{pool: pool}, nil
}

func (s *NeonStore) StoreSteps(ctx context.Context, steps int) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO steps (steps) VALUES ($1)", steps)
	return err
}
```

- [ ] **Step 4: Run to confirm tests pass**

```bash
go test ./gateways/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/gateways`

- [ ] **Step 5: Commit**

```bash
git add gateways/db.go gateways/db_test.go
git commit -m "feat: add NeonStore and update HealthStore interface"
```

---

## Task 5: Update controllers/steps.go — JSON parsing + context

**Files:**
- Modify: `controllers/steps.go`
- Modify: `controllers/steps_test.go`

The controller now:
- Takes a `StepStore` in addition to `StepLogger`
- `Handle` receives `ctx context.Context` as first arg
- Parses `{"steps": N}` — returns `ErrBadRequest`-wrapped error on failure
- Calls logger with raw bytes, then calls `StoreSteps(ctx, stepsCount)`

- [ ] **Step 1: Write the failing tests**

Overwrite `controllers/steps_test.go` with:

```go
package controllers_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/controllers"
	"github.com/rugpanov/ahri-health-bridge/utils"
)

type mockLogger struct {
	source string
	body   []byte
}

func (m *mockLogger) Log(source string, body []byte) {
	m.source = source
	m.body = body
}

type mockStore struct {
	steps int
	err   error
}

func (m *mockStore) StoreSteps(_ context.Context, steps int) error {
	m.steps = steps
	return m.err
}

func TestStepsController_Handle_CallsLoggerAndStore(t *testing.T) {
	logger := &mockLogger{}
	store := &mockStore{}
	ctrl := controllers.NewStepsController(logger, store)

	err := ctrl.Handle(context.Background(), []byte(`{"steps":500}`))

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if logger.source != "steps" {
		t.Errorf("expected source 'steps', got '%s'", logger.source)
	}
	if store.steps != 500 {
		t.Errorf("expected store to receive 500, got %d", store.steps)
	}
}

func TestStepsController_Handle_ZeroSteps(t *testing.T) {
	logger := &mockLogger{}
	store := &mockStore{}
	ctrl := controllers.NewStepsController(logger, store)

	err := ctrl.Handle(context.Background(), []byte(`{"steps":0}`))

	if err != nil {
		t.Errorf("unexpected error for zero steps: %v", err)
	}
	if store.steps != 0 {
		t.Errorf("expected 0 steps stored, got %d", store.steps)
	}
}

func TestStepsController_Handle_InvalidJSON(t *testing.T) {
	ctrl := controllers.NewStepsController(&mockLogger{}, &mockStore{})

	err := ctrl.Handle(context.Background(), []byte(`not json`))

	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !errors.Is(err, utils.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got: %v", err)
	}
}

func TestStepsController_Handle_MissingStepsField(t *testing.T) {
	ctrl := controllers.NewStepsController(&mockLogger{}, &mockStore{})

	err := ctrl.Handle(context.Background(), []byte(`{}`))

	if err == nil {
		t.Error("expected error for missing steps field")
	}
	if !errors.Is(err, utils.ErrBadRequest) {
		t.Errorf("expected ErrBadRequest, got: %v", err)
	}
}

func TestStepsController_Handle_StoreError(t *testing.T) {
	store := &mockStore{err: errors.New("db down")}
	ctrl := controllers.NewStepsController(&mockLogger{}, store)

	err := ctrl.Handle(context.Background(), []byte(`{"steps":100}`))

	if err == nil {
		t.Error("expected error when store fails")
	}
	if errors.Is(err, utils.ErrBadRequest) {
		t.Error("store error should not be wrapped as ErrBadRequest")
	}
}

```

- [ ] **Step 2: Run to confirm they fail**

```bash
go test ./controllers/...
```

Expected: compilation errors — `NewStepsController` wrong arity, `Handle` wrong signature

- [ ] **Step 3: Replace controllers/steps.go**

Overwrite `controllers/steps.go` with:

```go
package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

// StepLogger is the logging interface required by the steps controller.
type StepLogger interface {
	Log(source string, body []byte)
}

// StepStore is the storage interface required by the steps controller.
type StepStore interface {
	StoreSteps(ctx context.Context, steps int) error
}

type StepsController struct {
	logger StepLogger
	store  StepStore
}

func NewStepsController(logger StepLogger, store StepStore) *StepsController {
	return &StepsController{logger: logger, store: store}
}

func (c *StepsController) Handle(ctx context.Context, body []byte) error {
	var payload struct {
		Steps *int `json:"steps"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("%w: invalid JSON", utils.ErrBadRequest)
	}
	if payload.Steps == nil {
		return fmt.Errorf("%w: missing steps field", utils.ErrBadRequest)
	}
	c.logger.Log("steps", body)
	return c.store.StoreSteps(ctx, *payload.Steps)
}
```

- [ ] **Step 4: Run to confirm tests pass**

```bash
go test ./controllers/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/controllers`

- [ ] **Step 5: Commit**

```bash
git add controllers/steps.go controllers/steps_test.go
git commit -m "feat: parse steps JSON and wire store into controller"
```

---

## Task 6: Update handlers/steps.go — context + 400 vs 500

**Files:**
- Modify: `handlers/steps.go`
- Modify: `handlers/steps_test.go`

The handler's `StepsControllerI` interface adds `ctx context.Context`. The `ServeHTTP` method passes `r.Context()`. Errors wrapping `ErrBadRequest` return `400`; all others return `500`.

- [ ] **Step 1: Write the failing tests**

Overwrite `handlers/steps_test.go` with:

```go
package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/handlers"
	"github.com/rugpanov/ahri-health-bridge/utils"
)

type mockController struct {
	body []byte
	err  error
}

func (m *mockController) Handle(_ context.Context, body []byte) error {
	m.body = body
	return m.err
}

func TestStepsHandler_Success(t *testing.T) {
	mock := &mockController{}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health/steps", strings.NewReader(`{"steps":100}`))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != `{"status":"received"}` {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
	if string(mock.body) != `{"steps":100}` {
		t.Errorf("expected controller to receive body, got: %s", mock.body)
	}
}

func TestStepsHandler_EmptyBody(t *testing.T) {
	mock := &mockController{}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health/steps", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 even for empty body (controller decides validity), got %d", rr.Code)
	}
}

func TestStepsHandler_ValidationError_Returns400(t *testing.T) {
	mock := &mockController{err: fmt.Errorf("%w: invalid JSON", utils.ErrBadRequest)}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health/steps", strings.NewReader(`bad`))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestStepsHandler_StoreError_Returns500(t *testing.T) {
	mock := &mockController{err: errors.New("db down")}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health/steps", strings.NewReader(`{"steps":1}`))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
```

- [ ] **Step 2: Run to confirm they fail**

```bash
go test ./handlers/...
```

Expected: compilation error — `Handle` signature mismatch

- [ ] **Step 3: Replace handlers/steps.go**

Overwrite `handlers/steps.go` with:

```go
package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

// StepsControllerI is the controller interface required by the steps handler.
type StepsControllerI interface {
	Handle(ctx context.Context, body []byte) error
}

type StepsHandler struct {
	controller StepsControllerI
}

func NewStepsHandler(controller StepsControllerI) *StepsHandler {
	return &StepsHandler{controller: controller}
}

func (h *StepsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.controller.Handle(r.Context(), body); err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}
```

- [ ] **Step 4: Run to confirm tests pass**

```bash
go test ./handlers/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/handlers`

- [ ] **Step 5: Run full test suite**

```bash
go test ./...
```

Expected: all packages pass

- [ ] **Step 6: Commit**

```bash
git add handlers/steps.go handlers/steps_test.go
git commit -m "feat: add context and 400/500 error distinction to steps handler"
```

---

## Task 7: Update main.go + .env.example + smoke test

**Files:**
- Modify: `main.go`
- Modify: `.env.example`

- [ ] **Step 1: Add DATABASE_URL to .env.example**

Open `.env.example` and add the line:

```
DATABASE_URL=postgresql://user:pass@host/db?sslmode=require
```

Also add it to `.env` with the real Neon connection string.

- [ ] **Step 2: Replace main.go**

Overwrite `main.go` with:

```go
package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"

	"github.com/rugpanov/ahri-health-bridge/controllers"
	"github.com/rugpanov/ahri-health-bridge/gateways"
	"github.com/rugpanov/ahri-health-bridge/handlers"
	"github.com/rugpanov/ahri-health-bridge/utils"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "payloads.json"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	if err := runMigrations(databaseURL); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	ctx := context.Background()
	store, err := gateways.NewNeonStore(ctx, databaseURL)
	if err != nil {
		log.Fatalf("error creating store: %v", err)
	}

	logger, err := gateways.NewLogger(logFile)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}

	stepsCtrl := controllers.NewStepsController(logger, store)
	stepsHandler := handlers.NewStepsHandler(stepsCtrl)

	r := chi.NewRouter()
	r.Use(utils.APIKeyMiddleware(apiKey))
	r.Post("/health/steps", stepsHandler.ServeHTTP)

	fmt.Printf("ahri-health-bridge listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func runMigrations(databaseURL string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("creating migration source: %w", err)
	}
	migrateURL := strings.NewReplacer(
		"postgresql://", "pgx5://",
		"postgres://", "pgx5://",
	).Replace(databaseURL)
	m, err := migrate.NewWithSourceInstance("iofs", src, migrateURL)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Build to confirm it compiles**

```bash
go build .
```

Expected: no output, binary produced

- [ ] **Step 4: Run all tests**

```bash
go test ./...
```

Expected: all packages pass

- [ ] **Step 5: Start server and smoke test**

Ensure `.env` has `DATABASE_URL` set to the real Neon connection string.

Start server:
```bash
go run . &
SERVER_PID=$!
sleep 2
```

Test auth rejection:
```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/health/steps
```
Expected: `401`

Test invalid JSON (expect 400):
```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/health/steps \
  -H "X-API-Key: $(grep '^API_KEY' .env | cut -d= -f2)" \
  -d 'not json'
```
Expected: `400`

Test missing steps field (expect 400):
```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/health/steps \
  -H "X-API-Key: $(grep '^API_KEY' .env | cut -d= -f2)" \
  -H "Content-Type: application/json" \
  -d '{}'
```
Expected: `400`

Test valid request (expect 200 + DB insert):
```bash
curl -s -X POST http://localhost:8080/health/steps \
  -H "X-API-Key: $(grep '^API_KEY' .env | cut -d= -f2)" \
  -H "Content-Type: application/json" \
  -d '{"steps":1000}'
```
Expected: `{"status":"received"}`

Stop server:
```bash
kill $SERVER_PID 2>/dev/null
```

- [ ] **Step 6: Commit**

```bash
git add main.go .env.example
git commit -m "feat: add Neon storage, migrations, and DATABASE_URL config"
```
