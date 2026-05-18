package gateways

import "context"

// HealthStore is the interface for persisting health data.
// Implement a real version when Neon storage is ready.
type HealthStore interface {
	StoreSteps(ctx context.Context, body []byte) error
}

type NoopStore struct{}

func (n *NoopStore) StoreSteps(_ context.Context, _ []byte) error {
	return nil
}
