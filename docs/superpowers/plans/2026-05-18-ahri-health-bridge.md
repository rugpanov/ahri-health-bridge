# ahri-health-bridge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go HTTP service that receives Apple Health steps data from iOS Shortcuts, logs raw payloads to stdout and a local JSON file, and is ready to extend to additional health parameters.

**Architecture:** Chi router with auth middleware guarding a `/health/steps` POST endpoint. Requests flow: handler → controller → logger gateway. A no-op DB gateway stub defines the interface for future Neon storage. Config is loaded from `.env` at startup; missing `API_KEY` causes a hard exit.

**Tech Stack:** Go 1.22+, `go-chi/chi v5`, `joho/godotenv`

---

## Orchestration Model (Linear + Sub-agents)

Tasks 1–7 are tracked in Linear. A **main orchestrator agent** drives the loop:

```
for each Linear issue in "ahri-health-bridge" project (ordered by number):
  1. Set issue → In Progress
  2. Dispatch sub-agent with: task title, full task steps from this plan, repo path
  3. Sub-agent implements, runs tests, commits — reports back with commit SHA + test output
  4. Main agent verifies:
       - git log confirms the expected commit exists
       - go test ./... passes
       - changed files match the task's File Map
  5. If verification passes → set issue → Done, proceed to next
     If verification fails → leave issue In Progress, report diff to user before continuing
```

Tasks 1 and 4 have no inter-task dependencies and can be dispatched in parallel. All others must run in order (each task depends on the previous compile succeeding).

---

## File Map

| File | Responsibility |
|---|---|
| `main.go` | Entry point: load config, wire dependencies, start server |
| `utils/auth.go` | `APIKeyMiddleware` — validates `X-API-Key` header |
| `gateways/logger.go` | `Logger` struct — writes to stdout + appends JSON lines to file |
| `gateways/db.go` | `HealthStore` interface + `NoopStore` stub |
| `controllers/steps.go` | `StepsController` — calls logger, satisfies handler contract |
| `handlers/steps.go` | `StepsHandler` — reads body, calls controller, returns JSON |
| `utils/auth_test.go` | Tests for auth middleware |
| `gateways/logger_test.go` | Tests for logger file + stdout output |
| `controllers/steps_test.go` | Tests for controller with mock logger |
| `handlers/steps_test.go` | Tests for handler with mock controller |
| `.env.example` | Template for local secrets |

---

## Task 0: Set up Linear and create project issues

**Files:**
- Create: `.env.example`
- Create: `.env` (from `.env.example`, not committed)

This task is run once by a human (or the main orchestrator before dispatching sub-agents). It requires a Linear API key — follow Step 1 to get one if you don't have it.

- [ ] **Step 1: Get a Linear API key**

Go to [linear.app](https://linear.app) → Settings → API → Personal API keys → Create key.
Copy the key — it is only shown once.

If you do not have a Linear account, create one at [linear.app](https://linear.app) (free tier is sufficient).

- [ ] **Step 2: Create `.env.example` with all config fields**

Create `.env.example` in `tools/ahri-health-bridge/`:

```
PORT=8080
API_KEY=change-me
LOG_FILE=payloads.json
LINEAR_API_KEY=your-linear-api-key
```

Copy it to `.env` and fill in real values:

```bash
cp .env.example .env
```

Edit `.env`: set `API_KEY` to a random secret and `LINEAR_API_KEY` to the key from Step 1.

- [ ] **Step 3: Find your team ID**

```bash
curl -s -X POST https://api.linear.app/graphql \
  -H "Authorization: Bearer $(grep LINEAR_API_KEY .env | cut -d= -f2)" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ teams { nodes { id name } } }"}' | jq '.data.teams.nodes'
```

Expected: JSON array of teams. Note the `id` of the team you want to use (e.g. `"abc123"`).
If there are no teams, go to Linear → Settings → Workspace → Create team first.

- [ ] **Step 4: Get workflow state IDs for that team**

Replace `TEAM_ID` with the id from Step 3:

```bash
curl -s -X POST https://api.linear.app/graphql \
  -H "Authorization: Bearer $(grep LINEAR_API_KEY .env | cut -d= -f2)" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ workflowStates(filter: { team: { id: { eq: \"TEAM_ID\" } } }) { nodes { id name } } }"}' \
  | jq '.data.workflowStates.nodes'
```

Expected: states like `Todo`, `In Progress`, `Done`. Note the `id` for `Todo`.

- [ ] **Step 5: Create the project**

Replace `TEAM_ID`:

```bash
curl -s -X POST https://api.linear.app/graphql \
  -H "Authorization: Bearer $(grep LINEAR_API_KEY .env | cut -d= -f2)" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createProject(input: { name: \"ahri-health-bridge\", teamIds: [\"TEAM_ID\"] }) { project { id name } } }"
  }' | jq '.data.createProject.project'
```

Note the returned `id` — this is your `PROJECT_ID`.

- [ ] **Step 6: Create issues for Tasks 1–7**

Run once per task. Replace `TEAM_ID`, `PROJECT_ID`, and `TODO_STATE_ID` throughout.

```bash
LINEAR_API_KEY=$(grep LINEAR_API_KEY .env | cut -d= -f2)
TEAM_ID="your-team-id"
PROJECT_ID="your-project-id"
TODO_STATE_ID="your-todo-state-id"

create_issue() {
  local title="$1"
  local description="$2"
  curl -s -X POST https://api.linear.app/graphql \
    -H "Authorization: Bearer $LINEAR_API_KEY" \
    -H "Content-Type: application/json" \
    -d "{\"query\": \"mutation { createIssue(input: { teamId: \\\"$TEAM_ID\\\", projectId: \\\"$PROJECT_ID\\\", stateId: \\\"$TODO_STATE_ID\\\", title: \\\"$title\\\", description: \\\"$description\\\" }) { issue { id title } } }\"}" \
    | jq '.data.createIssue.issue'
}

create_issue "Task 1: Initialise Go module and install dependencies" "go mod init, install chi + godotenv, create .env.example. See plan Task 1."
create_issue "Task 2: Auth middleware" "X-API-Key middleware in utils/auth.go with tests. See plan Task 2."
create_issue "Task 3: Logger gateway" "stdout + file logger in gateways/logger.go with tests. See plan Task 3."
create_issue "Task 4: DB gateway stub" "HealthStore interface + NoopStore in gateways/db.go. See plan Task 4."
create_issue "Task 5: Steps controller" "StepsController in controllers/steps.go with tests. See plan Task 5."
create_issue "Task 6: Steps handler" "StepsHandler in handlers/steps.go with tests. See plan Task 6."
create_issue "Task 7: Wire up main.go and smoke test" "Assemble all components in main.go, run all tests, smoke test with curl. See plan Task 7."
```

Expected: 7 JSON objects printed, each with an `id` and `title`.

- [ ] **Step 7: Verify issues in Linear UI**

Open [linear.app](https://linear.app) and confirm 7 issues appear in the `ahri-health-bridge` project with status `Todo`.

- [ ] **Step 8: Commit**

```bash
git add .env.example
git commit -m "chore: add Linear setup and project issues for ahri-health-bridge"
```

---

## Task 1: Initialise Go module and install dependencies

**Files:**
- Create: `go.mod`

Prerequisite: Task 0 must be complete (`.env.example` and `.env` already exist).

- [ ] **Step 1: Initialise module**

```bash
cd tools/ahri-health-bridge
go mod init github.com/rugpanov/ahri-health-bridge
```

Expected: `go.mod` created with `module github.com/rugpanov/ahri-health-bridge`

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/go-chi/chi/v5
go get github.com/joho/godotenv
```

Expected: `go.sum` created, dependencies appear in `go.mod`

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: initialise Go module for ahri-health-bridge"
```

---

## Task 2: Auth middleware

**Files:**
- Create: `utils/auth.go`
- Create: `utils/auth_test.go`

- [ ] **Step 1: Write the failing tests**

Create `utils/auth_test.go`:

```go
package utils_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestAPIKeyMiddleware_ValidKey(t *testing.T) {
	handler := utils.APIKeyMiddleware("secret")(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", "secret")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_InvalidKey(t *testing.T) {
	handler := utils.APIKeyMiddleware("secret")(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("X-API-Key", "wrong")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAPIKeyMiddleware_MissingKey(t *testing.T) {
	handler := utils.APIKeyMiddleware("secret")(okHandler())
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./utils/...
```

Expected: compilation error — `utils.APIKeyMiddleware` undefined

- [ ] **Step 3: Implement auth middleware**

Create `utils/auth.go`:

```go
package utils

import "net/http"

func APIKeyMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-API-Key") != apiKey {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./utils/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/utils`

- [ ] **Step 5: Commit**

```bash
git add utils/auth.go utils/auth_test.go
git commit -m "feat: add API key auth middleware"
```

---

## Task 3: Logger gateway

**Files:**
- Create: `gateways/logger.go`
- Create: `gateways/logger_test.go`

- [ ] **Step 1: Write the failing tests**

Create `gateways/logger_test.go`:

```go
package gateways_test

import (
	"os"
	"strings"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/gateways"
)

func TestLogger_WritesToFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/payloads.json"

	logger, err := gateways.NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger error: %v", err)
	}

	logger.Log("steps", []byte(`{"count":1000}`))

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	if !strings.Contains(string(data), "steps") {
		t.Errorf("expected 'steps' in log file, got: %s", string(data))
	}
	if !strings.Contains(string(data), `{"count":1000}`) {
		t.Errorf("expected payload in log file, got: %s", string(data))
	}
}

func TestLogger_AppendsMultipleEntries(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/payloads.json"

	logger, err := gateways.NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger error: %v", err)
	}

	logger.Log("steps", []byte(`{"count":100}`))
	logger.Log("steps", []byte(`{"count":200}`))

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines in log file, got %d: %s", len(lines), string(data))
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./gateways/...
```

Expected: compilation error — `gateways.NewLogger` undefined

- [ ] **Step 3: Implement logger gateway**

Create `gateways/logger.go`:

```go
package gateways

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Logger struct {
	file *os.File
}

func NewLogger(path string) (*Logger, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

func (l *Logger) Log(source string, body []byte) {
	entry := map[string]any{
		"source":    source,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"payload":   string(body),
	}
	line, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("[%s] ERROR marshalling log entry: %v\n", source, err)
		return
	}
	fmt.Printf("[%s] %s\n", source, string(line))
	if _, err := l.file.Write(append(line, '\n')); err != nil {
		fmt.Printf("[%s] ERROR writing to log file: %v\n", source, err)
	}
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./gateways/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/gateways`

- [ ] **Step 5: Commit**

```bash
git add gateways/logger.go gateways/logger_test.go
git commit -m "feat: add logger gateway (stdout + file)"
```

---

## Task 4: DB gateway stub

**Files:**
- Create: `gateways/db.go`

No tests — this is a no-op interface stub that defines the contract for future Neon storage.

- [ ] **Step 1: Create DB gateway stub**

Create `gateways/db.go`:

```go
package gateways

import "context"

// HealthStore is the interface for persisting health data.
// Implement a real version when Neon storage is ready.
type HealthStore interface {
	StoreSteps(ctx context.Context, body []byte) error
}

type NoopStore struct{}

func (n *NoopStore) StoreSteps(_ context.Context, _ []byte) error {
	return nil
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./gateways/...
```

Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add gateways/db.go
git commit -m "feat: add HealthStore interface and NoopStore stub"
```

---

## Task 5: Steps controller

**Files:**
- Create: `controllers/steps.go`
- Create: `controllers/steps_test.go`

- [ ] **Step 1: Write the failing tests**

Create `controllers/steps_test.go`:

```go
package controllers_test

import (
	"testing"

	"github.com/rugpanov/ahri-health-bridge/controllers"
)

type mockLogger struct {
	source string
	body   []byte
}

func (m *mockLogger) Log(source string, body []byte) {
	m.source = source
	m.body = body
}

func TestStepsController_Handle_CallsLogger(t *testing.T) {
	mock := &mockLogger{}
	ctrl := controllers.NewStepsController(mock)

	body := []byte(`{"count":500}`)
	err := ctrl.Handle(body)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if mock.source != "steps" {
		t.Errorf("expected source 'steps', got '%s'", mock.source)
	}
	if string(mock.body) != string(body) {
		t.Errorf("expected body %s, got %s", body, mock.body)
	}
}

func TestStepsController_Handle_EmptyBody(t *testing.T) {
	mock := &mockLogger{}
	ctrl := controllers.NewStepsController(mock)

	err := ctrl.Handle([]byte{})

	if err != nil {
		t.Errorf("expected no error for empty body, got: %v", err)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./controllers/...
```

Expected: compilation error — `controllers.NewStepsController` undefined

- [ ] **Step 3: Implement steps controller**

Create `controllers/steps.go`:

```go
package controllers

// StepLogger is the logging interface required by the steps controller.
type StepLogger interface {
	Log(source string, body []byte)
}

type StepsController struct {
	logger StepLogger
}

func NewStepsController(logger StepLogger) *StepsController {
	return &StepsController{logger: logger}
}

func (c *StepsController) Handle(body []byte) error {
	c.logger.Log("steps", body)
	return nil
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./controllers/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/controllers`

- [ ] **Step 5: Commit**

```bash
git add controllers/steps.go controllers/steps_test.go
git commit -m "feat: add steps controller"
```

---

## Task 6: Steps handler

**Files:**
- Create: `handlers/steps.go`
- Create: `handlers/steps_test.go`

- [ ] **Step 1: Write the failing tests**

Create `handlers/steps_test.go`:

```go
package handlers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/handlers"
)

type mockController struct {
	body []byte
	err  error
}

func (m *mockController) Handle(body []byte) error {
	m.body = body
	return m.err
}

func TestStepsHandler_Success(t *testing.T) {
	mock := &mockController{}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health/steps", strings.NewReader(`{"count":100}`))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != `{"status":"received"}` {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
	if string(mock.body) != `{"count":100}` {
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
		t.Errorf("expected 200 even for empty body, got %d", rr.Code)
	}
}

func TestStepsHandler_ControllerError(t *testing.T) {
	mock := &mockController{err: errors.New("log failed")}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/health/steps", strings.NewReader(`{}`))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
go test ./handlers/...
```

Expected: compilation error — `handlers.NewStepsHandler` undefined

- [ ] **Step 3: Implement steps handler**

Create `handlers/steps.go`:

```go
package handlers

import (
	"io"
	"net/http"
)

// StepsControllerI is the controller interface required by the steps handler.
type StepsControllerI interface {
	Handle(body []byte) error
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

	if err := h.controller.Handle(body); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
go test ./handlers/...
```

Expected: `ok  github.com/rugpanov/ahri-health-bridge/handlers`

- [ ] **Step 5: Commit**

```bash
git add handlers/steps.go handlers/steps_test.go
git commit -m "feat: add steps handler"
```

---

## Task 7: Wire up main.go and smoke test

**Files:**
- Create: `main.go`

- [ ] **Step 1: Create main.go**

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/rugpanov/ahri-health-bridge/controllers"
	"github.com/rugpanov/ahri-health-bridge/gateways"
	"github.com/rugpanov/ahri-health-bridge/handlers"
	"github.com/rugpanov/ahri-health-bridge/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
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

	logger, err := gateways.NewLogger(logFile)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}

	stepsCtrl := controllers.NewStepsController(logger)
	stepsHandler := handlers.NewStepsHandler(stepsCtrl)

	r := chi.NewRouter()
	r.Use(utils.APIKeyMiddleware(apiKey))
	r.Post("/health/steps", stepsHandler.ServeHTTP)

	fmt.Printf("ahri-health-bridge listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
```

- [ ] **Step 2: Build to confirm it compiles**

```bash
go build .
```

Expected: no output, binary produced

- [ ] **Step 3: Run all tests**

```bash
go test ./...
```

Expected: all packages pass

- [ ] **Step 4: Start the server and smoke test**

In one terminal:
```bash
go run .
```

Expected output: `ahri-health-bridge listening on :8080`

In a second terminal — test auth rejection:
```bash
curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/health/steps
```
Expected: `401`

Test valid request:
```bash
curl -s -X POST http://localhost:8080/health/steps \
  -H "X-API-Key: $(grep API_KEY .env | cut -d= -f2)" \
  -H "Content-Type: application/json" \
  -d '{"count":1000,"source":"iPhone","timestamp":"2026-05-18T10:00:00Z"}'
```
Expected: `{"status":"received"}` printed, entry visible in stdout and in `payloads.json`

- [ ] **Step 5: Commit**

```bash
git add main.go
git commit -m "feat: wire up ahri-health-bridge server"
```
