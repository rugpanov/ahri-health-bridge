package handlers

import (
	"io"
	"net/http"
)

// StepsControllerI is the controller interface required by the steps handler.
type StepsControllerI interface {
	Handle(body []byte) error
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

	if err := h.controller.Handle(body); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"received"}`))
}
