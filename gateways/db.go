package gateways

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rugpanov/ahri-health-bridge/controllers"
)

type NoopStore struct{}

func (n *NoopStore) StoreSteps(_ context.Context, _ int) error {
	return nil
}

func (n *NoopStore) GetStepsByDay(_ context.Context) ([]controllers.DailyStepsRecord, error) {
	return nil, nil
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

func (s *NeonStore) GetStepsByDay(ctx context.Context) ([]controllers.DailyStepsRecord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT DATE(received_at) AS day, SUM(steps) AS total_steps
		FROM steps
		GROUP BY day
		ORDER BY day ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []controllers.DailyStepsRecord
	for rows.Next() {
		var r controllers.DailyStepsRecord
		if err := rows.Scan(&r.Date, &r.Steps); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}
