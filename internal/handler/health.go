package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.Error("writing json response",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)
	}
}
