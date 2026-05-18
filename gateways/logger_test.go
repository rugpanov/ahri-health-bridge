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
	if !strings.Contains(string(data), `"payload":"{`) {
		t.Errorf("expected payload string in log file, got: %s", string(data))
	}
	if !strings.Contains(string(data), `count`) {
		t.Errorf("expected count in payload, got: %s", string(data))
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
