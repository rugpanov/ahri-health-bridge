package handlers

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/rugpanov/ahri-health-bridge/utils"
)

type StepsControllerI interface {
	Handle(ctx context.Context, body []byte) error
}

type StepsHandler struct {
	controller StepsControllerI
}

func NewStepsHandler(controller StepsControllerI) *StepsHandler {
	return &StepsHandler{controller: controller}
}

func (h *StepsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := h.controller.Handle(r.Context(), body); err != nil {
		if errors.Is(err, utils.ErrBadRequest) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}
