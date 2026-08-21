package handlers

import (
	"net/http"
	"strings"

	"MaintControl/internal/middleware"
	"MaintControl/internal/services"
)

type ProductionLineHandler struct {
	lines *services.ProductionLineService
}

func NewProductionLineHandler(lines *services.ProductionLineService) *ProductionLineHandler {
	return &ProductionLineHandler{lines: lines}
}

func (h *ProductionLineHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if r.URL.Path == "/production-lines" || r.URL.Path == "/production-lines/" {
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

	parts := splitPath(r.URL.Path, "/production-lines/")
	if len(parts) == 0 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	lineID := parts[0]
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.get(w, r, userID, lineID)
		case http.MethodPut:
			h.update(w, r, userID, lineID)
		case http.MethodDelete:
			h.delete(w, r, userID, lineID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 2 && parts[1] == "machines" {
		switch r.Method {
		case http.MethodPost:
			h.addMachine(w, r, userID, lineID)
		case http.MethodGet:
			h.listMachines(w, r, userID, lineID)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 3 && parts[1] == "machines" && r.Method == http.MethodDelete {
		h.removeMachine(w, r, userID, lineID, parts[2])
		return
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (h *ProductionLineHandler) create(w http.ResponseWriter, r *http.Request, userID string) {
	var request struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	line, err := h.lines.Create(r.Context(), userID, request.Name)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, line)
}

func (h *ProductionLineHandler) list(w http.ResponseWriter, r *http.Request, userID string) {
	lines, err := h.lines.List(r.Context(), userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, lines)
}

func (h *ProductionLineHandler) get(w http.ResponseWriter, r *http.Request, userID, id string) {
	line, machines, err := h.lines.Get(r.Context(), userID, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"production_line": line,
		"machines":        machines,
	})
}

func (h *ProductionLineHandler) update(w http.ResponseWriter, r *http.Request, userID, id string) {
	var request struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	line, err := h.lines.Update(r.Context(), userID, id, request.Name)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, line)
}

func (h *ProductionLineHandler) delete(w http.ResponseWriter, r *http.Request, userID, id string) {
	if err := h.lines.Delete(r.Context(), userID, id); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductionLineHandler) addMachine(w http.ResponseWriter, r *http.Request, userID, lineID string) {
	var request struct {
		MachineID string `json:"machine_id"`
		Position  int    `json:"position"`
	}
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	link, err := h.lines.AddMachine(r.Context(), userID, lineID, request.MachineID, request.Position)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, link)
}

func (h *ProductionLineHandler) removeMachine(w http.ResponseWriter, r *http.Request, userID, lineID, machineID string) {
	if err := h.lines.RemoveMachine(r.Context(), userID, lineID, machineID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductionLineHandler) listMachines(w http.ResponseWriter, r *http.Request, userID, lineID string) {
	machines, err := h.lines.ListMachines(r.Context(), userID, lineID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, machines)
}

func splitPath(path, prefix string) []string {
	trimmed := strings.Trim(strings.TrimPrefix(path, prefix), "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}
