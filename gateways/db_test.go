package gateways_test

import (
	"context"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/gateways"
)

func TestNoopStore_StoreSteps(t *testing.T) {
	store := &gateways.NoopStore{}
	err := store.StoreSteps(context.Background(), 100)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestNoopStore_StoreSteps_Zero(t *testing.T) {
	store := &gateways.NoopStore{}
	err := store.StoreSteps(context.Background(), 0)
	if err != nil {
		t.Errorf("expected no error for zero steps, got: %v", err)
	}
}
