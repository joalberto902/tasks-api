package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, r *http.Request, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writing json response",
			"error", err,
			"method", r.Method,
			"path", r.URL.Path)
	}
}

func writeError(w http.ResponseWriter, r *http.Request, code int, msg string) {
	v := map[string]string{"error": msg}

	writeJSON(w, r, code, v)
}
