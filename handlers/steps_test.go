package handlers_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rugpanov/ahri-health-bridge/controllers"
	"github.com/rugpanov/ahri-health-bridge/handlers"
	"github.com/rugpanov/ahri-health-bridge/utils"
)

type mockController struct {
	body     []byte
	err      error
	daily    []controllers.DailyStepsResult
	dailyErr error
}

func (m *mockController) Handle(_ context.Context, body []byte) error {
	m.body = body
	return m.err
}

func (m *mockController) GetByDay(_ context.Context) ([]controllers.DailyStepsResult, error) {
	return m.daily, m.dailyErr
}

func TestStepsHandler_GetByDay_Success(t *testing.T) {
	date := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	_ = date
	mock := &mockController{daily: []controllers.DailyStepsResult{
		{Date: "2026-01-15", Steps: 8000},
	}}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/health/steps/daily", nil)
	rr := httptest.NewRecorder()

	h.GetByDayServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	want := `[{"date":"2026-01-15","steps":8000}]` + "\n"
	if rr.Body.String() != want {
		t.Errorf("unexpected body: %s", rr.Body.String())
	}
}

func TestStepsHandler_GetByDay_Empty(t *testing.T) {
	mock := &mockController{daily: nil}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/health/steps/daily", nil)
	rr := httptest.NewRecorder()

	h.GetByDayServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "[]\n" {
		t.Errorf("expected empty array, got: %s", rr.Body.String())
	}
}

func TestStepsHandler_GetByDay_StoreError(t *testing.T) {
	mock := &mockController{dailyErr: errors.New("db down")}
	h := handlers.NewStepsHandler(mock)

	req := httptest.NewRequest(http.MethodGet, "/health/steps/daily", nil)
	rr := httptest.NewRecorder()

	h.GetByDayServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
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
