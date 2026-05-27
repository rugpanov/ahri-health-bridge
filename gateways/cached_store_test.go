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
	rows     []controllers.DailyStepsRecord
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

func (f *fakeStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsRecord, error) {
	return f.rows, nil
}

func TestNewCachedStore_PopulatesFromInner(t *testing.T) {
	date1, _ := time.Parse("2006-01-02", "2026-01-01")
	date2, _ := time.Parse("2006-01-02", "2026-01-02")
	inner := &fakeStore{
		rows: []controllers.DailyStepsRecord{
			{Date: date1, Steps: 5000},
			{Date: date2, Steps: 8000},
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
	if got[0].Date.Format("2006-01-02") != "2026-01-01" || got[0].Steps != 5000 {
		t.Errorf("unexpected first record: %+v", got[0])
	}
	if got[1].Date.Format("2006-01-02") != "2026-01-02" || got[1].Steps != 8000 {
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
func (f *failingGetStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsRecord, error) {
	return nil, f.err
}

func TestGetStepsByDay_ReturnsCopy(t *testing.T) {
	date1, _ := time.Parse("2006-01-02", "2026-01-01")
	inner := &fakeStore{
		rows: []controllers.DailyStepsRecord{{Date: date1, Steps: 5000}},
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
	todayTime, _ := time.Parse("2006-01-02", today)
	inner := &fakeStore{
		rows: []controllers.DailyStepsRecord{
			{Date: todayTime, Steps: 3000},
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
	date1, _ := time.Parse("2006-01-02", "2026-01-01")
	inner := &fakeStore{
		rows: []controllers.DailyStepsRecord{
			{Date: date1, Steps: 5000},
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
	todayTime, _ := time.Parse("2006-01-02", today)
	inner := &fakeStore{
		rows:     []controllers.DailyStepsRecord{{Date: todayTime, Steps: 3000}},
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
