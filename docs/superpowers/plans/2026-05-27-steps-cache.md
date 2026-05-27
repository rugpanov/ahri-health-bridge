# Steps Read Cache Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an in-memory `CachedStore` decorator that serves `GET /health/steps/daily` reads from cache and keeps the cache current on every write.

**Architecture:** `CachedStore` in `gateways/` wraps any `controllers.StepStore`. The constructor populates the cache once from the inner store. `StoreSteps` writes to the inner store first, then updates the cache unconditionally (steps are cumulative — last write per day is always the highest). `GetStepsByDay` returns a copy of the cache under a read lock.

**Tech Stack:** Go 1.26, `sync.RWMutex`, `pgx/v5`, Chi router.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `gateways/cached_store.go` | Create | `CachedStore` struct, constructor, `StoreSteps`, `GetStepsByDay` |
| `gateways/cached_store_test.go` | Create | Unit tests using a fake inner store — no DB |
| `main.go` | Modify (lines ~56-60) | Wrap `NeonStore` with `NewCachedStore` |

---

### Task 1: `CachedStore` constructor with startup population

**Files:**
- Create: `gateways/cached_store.go`
- Create: `gateways/cached_store_test.go`

- [ ] **Step 1: Write the failing tests**

Create `gateways/cached_store_test.go`:

```go
package gateways_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rugpanov/ahri-health-bridge/controllers"
	"github.com/rugpanov/ahri-health-bridge/gateways"
)

// fakeStore is a controllable in-memory StepStore for testing.
type fakeStore struct {
	rows     []controllers.DailyStepsResult
	storeErr error
	stored   []int
}

func (f *fakeStore) StoreSteps(_ context.Context, steps int) error {
	if f.storeErr != nil {
		return f.storeErr
	}
	f.stored = append(f.stored, steps)
	return nil
}

func (f *fakeStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsResult, error) {
	return f.rows, nil
}

func TestNewCachedStore_PopulatesFromInner(t *testing.T) {
	inner := &fakeStore{
		rows: []controllers.DailyStepsResult{
			{Date: "2026-01-01", Steps: 5000},
			{Date: "2026-01-02", Steps: 8000},
		},
	}
	store, err := gateways.NewCachedStore(context.Background(), inner)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := store.GetStepsByDay(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
	if got[0].Date != "2026-01-01" || got[0].Steps != 5000 {
		t.Errorf("unexpected first record: %+v", got[0])
	}
	if got[1].Date != "2026-01-02" || got[1].Steps != 8000 {
		t.Errorf("unexpected second record: %+v", got[1])
	}
}

func TestNewCachedStore_FailsIfInnerFails(t *testing.T) {
	failing := &failingGetStore{err: errors.New("db down")}
	_, err := gateways.NewCachedStore(context.Background(), failing)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// failingGetStore returns an error from GetStepsByDay.
type failingGetStore struct {
	err error
}

func (f *failingGetStore) StoreSteps(_ context.Context, _ int) error { return nil }
func (f *failingGetStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsResult, error) {
	return nil, f.err
}

func TestGetStepsByDay_ReturnsCopy(t *testing.T) {
	inner := &fakeStore{
		rows: []controllers.DailyStepsResult{{Date: "2026-01-01", Steps: 5000}},
	}
	store, _ := gateways.NewCachedStore(context.Background(), inner)
	got, _ := store.GetStepsByDay(context.Background())
	got[0].Steps = 99999 // mutate the returned slice
	got2, _ := store.GetStepsByDay(context.Background())
	if got2[0].Steps == 99999 {
		t.Error("mutation of returned slice affected internal cache")
	}
}

func TestStoreSteps_UpdatesTodaysEntry(t *testing.T) {
	today := time.Now().UTC().Format("2006-01-02")
	inner := &fakeStore{
		rows: []controllers.DailyStepsResult{
			{Date: today, Steps: 3000},
		},
	}
	store, _ := gateways.NewCachedStore(context.Background(), inner)
	if err := store.StoreSteps(context.Background(), 7500); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := store.GetStepsByDay(context.Background())
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}
	if got[0].Steps != 7500 {
		t.Errorf("expected steps=7500, got %d", got[0].Steps)
	}
}

func TestStoreSteps_AppendsNewDay(t *testing.T) {
	inner := &fakeStore{
		rows: []controllers.DailyStepsResult{
			{Date: "2026-01-01", Steps: 5000},
		},
	}
	store, _ := gateways.NewCachedStore(context.Background(), inner)
	if err := store.StoreSteps(context.Background(), 4000); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, _ := store.GetStepsByDay(context.Background())
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
}

func TestStoreSteps_DoesNotUpdateCacheOnInnerError(t *testing.T) {
	today := time.Now().UTC().Format("2006-01-02")
	inner := &fakeStore{
		rows:     []controllers.DailyStepsResult{{Date: today, Steps: 3000}},
		storeErr: errors.New("db write failed"),
	}
	store, _ := gateways.NewCachedStore(context.Background(), inner)
	err := store.StoreSteps(context.Background(), 9999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	got, _ := store.GetStepsByDay(context.Background())
	if got[0].Steps != 3000 {
		t.Errorf("cache was updated despite inner error, got steps=%d", got[0].Steps)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

```bash
cd tools/ahri-health-bridge && go test ./gateways/... -run TestNewCachedStore -v
```

Expected: compile error — `gateways.NewCachedStore` undefined.

- [ ] **Step 3: Create `gateways/cached_store.go` with struct and constructor**

```go
package gateways

import (
	"context"
	"sync"
	"time"

	"github.com/rugpanov/ahri-health-bridge/controllers"
)

type CachedStore struct {
	inner controllers.StepStore
	mu    sync.RWMutex
	cache []controllers.DailyStepsResult
}

func NewCachedStore(ctx context.Context, inner controllers.StepStore) (*CachedStore, error) {
	rows, err := inner.GetStepsByDay(ctx)
	if err != nil {
		return nil, err
	}
	return &CachedStore{inner: inner, cache: rows}, nil
}

func (c *CachedStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsResult, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]controllers.DailyStepsResult, len(c.cache))
	copy(result, c.cache)
	return result, nil
}

func (c *CachedStore) StoreSteps(ctx context.Context, steps int) error {
	if err := c.inner.StoreSteps(ctx, steps); err != nil {
		return err
	}
	today := time.Now().UTC().Format("2006-01-02")
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, r := range c.cache {
		if r.Date == today {
			c.cache[i].Steps = steps
			return nil
		}
	}
	c.cache = append(c.cache, controllers.DailyStepsResult{Date: today, Steps: steps})
	return nil
}
```

- [ ] **Step 4: Run all tests and verify they pass**

```bash
cd tools/ahri-health-bridge && go test ./gateways/... -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
cd tools/ahri-health-bridge && git add gateways/cached_store.go gateways/cached_store_test.go
git commit -m "feat: add CachedStore decorator for steps read cache"
```

---

### Task 2: Wire `CachedStore` in `main.go`

**Files:**
- Modify: `main.go` (~line 57)

- [ ] **Step 1: Replace the store wiring**

In `main.go`, replace the block that creates the store:

```go
// Before:
store, err := gateways.NewNeonStore(ctx, databaseURL)
if err != nil {
    log.Fatalf("error creating store: %v", err)
}
```

```go
// After:
neon, err := gateways.NewNeonStore(ctx, databaseURL)
if err != nil {
    log.Fatalf("error creating store: %v", err)
}
store, err := gateways.NewCachedStore(ctx, neon)
if err != nil {
    log.Fatalf("error creating cache: %v", err)
}
```

- [ ] **Step 2: Verify the project compiles and all tests pass**

```bash
cd tools/ahri-health-bridge && go build ./... && go test ./...
```

Expected: no errors, all tests PASS.

- [ ] **Step 3: Commit**

```bash
cd tools/ahri-health-bridge && git add main.go
git commit -m "feat: wrap NeonStore with CachedStore in main"
```
