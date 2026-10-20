package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joalberto902/tasks-api/internal/handler"
	"github.com/joalberto902/tasks-api/internal/repository"
)

const (
	port           = ":8080"
	healthEndpoint = "/health"
	tasksEndpoint  = "/tasks"
	idPath         = "/{id}"
)

func main() {
	repo := repository.NewTaskRepository()
	taskHandler := handler.NewTaskHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+healthEndpoint, handler.HealthHandler)
	mux.HandleFunc("GET "+tasksEndpoint, taskHandler.List)
	mux.HandleFunc("GET "+tasksEndpoint+idPath, taskHandler.Get)
	mux.HandleFunc("POST "+tasksEndpoint, taskHandler.Create)

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
