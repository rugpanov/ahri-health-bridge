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
	cache []controllers.DailyStepsRecord
}

func NewCachedStore(ctx context.Context, inner controllers.StepStore) (*CachedStore, error) {
	rows, err := inner.GetStepsByDay(ctx)
	if err != nil {
		return nil, err
	}
	return &CachedStore{inner: inner, cache: rows}, nil
}

func (c *CachedStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsRecord, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]controllers.DailyStepsRecord, len(c.cache))
	copy(result, c.cache)
	return result, nil
}

func (c *CachedStore) StoreSteps(ctx context.Context, steps int) error {
	if err := c.inner.StoreSteps(ctx, steps); err != nil {
		return err
	}
	today := time.Now().UTC().Format("2006-01-02")
	todayTime, _ := time.Parse("2006-01-02", today)
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, r := range c.cache {
		if r.Date.Format("2006-01-02") == today {
			c.cache[i].Steps = steps
			return nil
		}
	}
	c.cache = append(c.cache, controllers.DailyStepsRecord{Date: todayTime, Steps: steps})
	return nil
}
