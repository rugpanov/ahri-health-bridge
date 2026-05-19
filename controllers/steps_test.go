package controllers_test

import (
	"context"
	"errors"
	"testing"
	"time"

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
	steps    int
	err      error
	daily    []controllers.DailyStepsRecord
	dailyErr error
}

func (m *mockStore) StoreSteps(_ context.Context, steps int) error {
	m.steps = steps
	return m.err
}

func (m *mockStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsRecord, error) {
	return m.daily, m.dailyErr
}

func TestStepsController_GetByDay(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	store := &mockStore{daily: []controllers.DailyStepsRecord{{Date: date, Steps: 8000}}}
	ctrl := controllers.NewStepsController(&mockLogger{}, store)

	results, err := ctrl.GetByDay(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Date != "2026-01-15" {
		t.Errorf("expected date '2026-01-15', got '%s'", results[0].Date)
	}
	if results[0].Steps != 8000 {
		t.Errorf("expected 8000 steps, got %d", results[0].Steps)
	}
}

func TestStepsController_GetByDay_StoreError(t *testing.T) {
	store := &mockStore{dailyErr: errors.New("db down")}
	ctrl := controllers.NewStepsController(&mockLogger{}, store)

	_, err := ctrl.GetByDay(context.Background())

	if err == nil {
		t.Error("expected error when store fails")
	}
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
