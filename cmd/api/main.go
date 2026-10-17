package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joalberto902/tasks-api/internal/handler"
)

const (
	port           = ":8080"
	healthEndpoint = "/health"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+healthEndpoint, handler.HealthHandler)

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
