package gateways

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthStore interface {
	StoreSteps(ctx context.Context, steps int) error
}

type NoopStore struct{}

func (n *NoopStore) StoreSteps(_ context.Context, _ int) error {
	return nil
}

type NeonStore struct {
	pool *pgxpool.Pool
}

func NewNeonStore(ctx context.Context, connString string) (*NeonStore, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &NeonStore{pool: pool}, nil
}

func (s *NeonStore) StoreSteps(ctx context.Context, steps int) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO steps (steps) VALUES ($1)", steps)
	return err
}
