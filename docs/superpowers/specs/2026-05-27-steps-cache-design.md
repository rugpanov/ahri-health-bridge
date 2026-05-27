# Steps Read Cache — Design Spec

**Date:** 2026-05-27
**Status:** Approved

## Problem

Every `GET /health/steps/daily` request hits Neon with a `SELECT … GROUP BY day` query. Given that writes are infrequent (iOS Shortcut, at most a few times per day) and the result set is small (one row per day of history), this is unnecessary latency and DB load.

## Solution

An in-memory cache that is populated once on startup and kept current on every write. Reads are always served from the cache — no DB round-trip.

## Key Observation

Steps are cumulative: the iOS Shortcut sends the running daily total each time it fires. The last value written each day is always the highest. This means the cache update on write is unconditional — no `MAX` comparison needed, just replace today's entry with the new value.

## Architecture

`CachedStore` is a decorator in `gateways/` that wraps any `controllers.StepStore`. It implements the same interface, so no changes are needed in `controllers/` or `handlers/`.

```
main.go
  └─ NewNeonStore(ctx, dsn)       → inner store (DB)
       └─ NewCachedStore(ctx, inner) → wraps inner, populates cache on startup
            └─ passed to NewStepsController(logger, store)
```

### Struct

```go
type CachedStore struct {
    inner controllers.StepStore
    mu    sync.RWMutex
    cache []controllers.DailyStepsResult
}
```

### Constructor

`NewCachedStore(ctx, inner)` calls `inner.GetStepsByDay(ctx)` once to populate the cache. Returns an error if that call fails — the service fails fast rather than starting with a silently empty cache.

## Data Flow

### Writes — `StoreSteps(ctx, steps)`

1. Call `inner.StoreSteps(ctx, steps)` — if this fails, return the error and do **not** update the cache.
2. Acquire write lock.
3. Find today's entry in the cache (match on `date == today`).
   - Found: replace its `Steps` value unconditionally.
   - Not found: append a new `DailyStepsResult{Date: today, Steps: steps}`.
4. Release write lock.

### Reads — `GetStepsByDay(ctx)`

1. Acquire read lock.
2. Copy the slice.
3. Release read lock.
4. Return the copy.

The copy prevents callers from holding a reference to the internal slice.

## Error Handling

| Scenario | Behaviour |
|---|---|
| Startup DB query fails | `NewCachedStore` returns error; service exits |
| DB write fails | Cache not updated; error returned to caller |
| DB write succeeds, cache update panics | Not expected; cache update is pure in-memory logic |

## Testing

- Unit-test `CachedStore` using `NoopStore` (already exists) as the inner store — no DB needed.
- Test startup population, write-then-read round-trip, today's entry replacement, and new-day append.
- Existing handler/controller tests are unaffected.

## Files Changed

| File | Change |
|---|---|
| `gateways/cached_store.go` | New file — `CachedStore` implementation |
| `gateways/cached_store_test.go` | New file — unit tests |
| `main.go` | Wrap `NeonStore` with `NewCachedStore` |
