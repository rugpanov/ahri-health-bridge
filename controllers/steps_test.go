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
