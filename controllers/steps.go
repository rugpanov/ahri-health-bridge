package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

type StepLogger interface {
	Log(source string, body []byte)
}

type DailyStepsRecord struct {
	Date  time.Time
	Steps int
}

type StepStore interface {
	StoreSteps(ctx context.Context, steps int) error
	GetStepsByDay(ctx context.Context) ([]DailyStepsRecord, error)
}

type DailyStepsResult struct {
	Date  string `json:"date"`
	Steps int    `json:"steps"`
}

type StepsController struct {
	logger StepLogger
	store  StepStore
}

func NewStepsController(logger StepLogger, store StepStore) *StepsController {
	return &StepsController{logger: logger, store: store}
}

func (c *StepsController) Handle(ctx context.Context, body []byte) error {
	var payload struct {
		Steps *int `json:"steps"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("%w: invalid JSON", utils.ErrBadRequest)
	}
	if payload.Steps == nil {
		return fmt.Errorf("%w: missing steps field", utils.ErrBadRequest)
	}
	c.logger.Log("steps", body)
	return c.store.StoreSteps(ctx, *payload.Steps)
}

func (c *StepsController) GetByDay(ctx context.Context) ([]DailyStepsResult, error) {
	rows, err := c.store.GetStepsByDay(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]DailyStepsResult, len(rows))
	for i, r := range rows {
		result[i] = DailyStepsResult{
			Date:  r.Date.Format("2006-01-02"),
			Steps: r.Steps,
		}
	}
	return result, nil
}
