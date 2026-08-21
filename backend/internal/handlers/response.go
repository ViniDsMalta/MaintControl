package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"MaintControl/internal/services"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func statusFromError(err error) (int, string) {
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid credentials"
	case errors.Is(err, services.ErrUnauthorized):
		return http.StatusUnauthorized, "unauthorized"
	case errors.Is(err, services.ErrDuplicateEmail):
		return http.StatusConflict, "email already exists"
	case errors.Is(err, services.ErrDuplicateAssociation):
		return http.StatusConflict, "machine already belongs to this production line"
	case errors.Is(err, services.ErrDuplicatePosition):
		return http.StatusConflict, "position already in use"
	case errors.Is(err, services.ErrConflict):
		return http.StatusConflict, "conflict"
	case errors.Is(err, services.ErrNotFound):
		return http.StatusNotFound, "not found"
	case errors.Is(err, services.ErrInvalidInput):
		return http.StatusBadRequest, "invalid input"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}

func handleServiceError(w http.ResponseWriter, err error) {
	status, message := statusFromError(err)
	writeError(w, status, message)
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
