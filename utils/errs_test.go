package utils_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

func TestErrBadRequest_IsWrappable(t *testing.T) {
	wrapped := fmt.Errorf("%w: invalid JSON", utils.ErrBadRequest)

	if !errors.Is(wrapped, utils.ErrBadRequest) {
		t.Error("expected errors.Is to match ErrBadRequest through wrapping")
	}
}
