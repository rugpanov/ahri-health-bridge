package controllers

// StepLogger is the logging interface required by the steps controller.
type StepLogger interface {
	Log(source string, body []byte)
}

type StepsController struct {
	logger StepLogger
}

func NewStepsController(logger StepLogger) *StepsController {
	return &StepsController{logger: logger}
}

func (c *StepsController) Handle(body []byte) error {
	c.logger.Log("steps", body)
	return nil
}
