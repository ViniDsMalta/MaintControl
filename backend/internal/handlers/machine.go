package handlers

import (
	"net/http"
	"strings"

	"MaintControl/internal/middleware"
	"MaintControl/internal/services"
)

type MachineHandler struct {
	machines *services.MachineService
}

func NewMachineHandler(machines *services.MachineService) *MachineHandler {
	return &MachineHandler{machines: machines}
}

func (h *MachineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/machines/")
	if r.URL.Path == "/machines" || r.URL.Path == "/machines/" {
		switch r.Method {
		case http.MethodPost:
			h.create(w, r, userID)
		case http.MethodGet:
			h.list(w, r, userID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.get(w, r, userID, id)
	case http.MethodPut:
		h.update(w, r, userID, id)
	case http.MethodDelete:
		h.delete(w, r, userID, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *MachineHandler) create(w http.ResponseWriter, r *http.Request, userID string) {
	var request struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	machine, err := h.machines.Create(r.Context(), userID, request.Name, request.Type)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, machine)
}

func (h *MachineHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	machines, err := h.machines.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, machines)
}

func (h *MachineHandler) get(w http.ResponseWriter, r *http.Request, userID, id string) {
	machine, err := h.machines.Get(r.Context(), userID, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, machine)
}

func (h *MachineHandler) update(w http.ResponseWriter, r *http.Request, userID, id string) {
	var request struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	machine, err := h.machines.Update(r.Context(), userID, id, request.Name, request.Type)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, machine)
}

func (h *MachineHandler) delete(w http.ResponseWriter, r *http.Request, userID, id string) {
	if err := h.machines.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
