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
