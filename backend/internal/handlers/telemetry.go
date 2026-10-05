package handlers

import (
	"net/http"
	"strings"

	"MaintControl/internal/models"
	"MaintControl/internal/services"
)

type TelemetryHandler struct {
	telemetry *services.TelemetryService
}

func NewTelemetryHandler(telemetry *services.TelemetryService) *TelemetryHandler {
	return &TelemetryHandler{telemetry: telemetry}
}

func (h *TelemetryHandler) Receive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var reading models.TelemetryReading
	if err := decodeJSON(r, &reading); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
	result, err := h.telemetry.Receive(r.Context(), apiKey, reading)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	status := http.StatusCreated
	if !result.PredictionPerformed {
		status = http.StatusAccepted
	}
	writeJSON(w, status, result)
}
