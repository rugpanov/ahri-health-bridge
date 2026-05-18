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
