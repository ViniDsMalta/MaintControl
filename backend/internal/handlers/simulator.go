package handlers

import (
	"net/http"

	"MaintControl/internal/models"
	"MaintControl/internal/services"
)

type SimulatorHandler struct {
	simulator *services.SimulatorService
	telemetry *services.TelemetryService
}

func NewSimulatorHandler(simulator *services.SimulatorService, telemetry *services.TelemetryService) *SimulatorHandler {
	return &SimulatorHandler{simulator: simulator, telemetry: telemetry}
}

func (h *SimulatorHandler) ListMachines(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	machines, err := h.simulator.ListMachines(r.Context(), r.Header.Get("X-Simulator-Token"))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, machines)
}

func (h *SimulatorHandler) ReceiveTelemetry(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var request struct {
		MachineID string                  `json:"machine_id"`
		Reading   models.TelemetryReading `json:"reading"`
	}
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	apiKey, err := h.simulator.MachineAPIKey(r.Context(), r.Header.Get("X-Simulator-Token"), request.MachineID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	result, err := h.telemetry.Receive(r.Context(), apiKey, request.Reading)
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
