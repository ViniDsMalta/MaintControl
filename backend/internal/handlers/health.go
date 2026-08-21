package handlers

import (
	"MaintControl/internal/database"
	"context"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	response := map[string]string{"status": "ok"}
	if database.DB != nil {
		if err := database.DB.Ping(context.Background()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":   "error",
				"database": "unavailable",
			})
			return
		}
		response["database"] = "ok"
	}

	writeJSON(w, http.StatusOK, response)
}
