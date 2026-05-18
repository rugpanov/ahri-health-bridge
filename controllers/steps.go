package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

type StepLogger interface {
	Log(source string, body []byte)
}

type StepStore interface {
	StoreSteps(ctx context.Context, steps int) error
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
