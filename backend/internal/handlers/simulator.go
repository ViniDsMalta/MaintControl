package handlers

import (
	"net/http"

	"MaintControl/internal/services"
)

type SimulatorHandler struct {
	simulator *services.SimulatorService
}

func NewSimulatorHandler(simulator *services.SimulatorService) *SimulatorHandler {
	return &SimulatorHandler{simulator: simulator}
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
