package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

const (
	port           = ":8080"
	healthEndpoint = "/health"
)

type healthStatusResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(healthStatusResponse{Status: "ok"})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+healthEndpoint, healthHandler)

	server := http.Server{
		Addr:              port,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Printf("Serving on port: %s\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}

}
