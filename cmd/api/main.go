package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joalberto902/tasks-api/internal/config"
	"github.com/joalberto902/tasks-api/internal/handler"
	"github.com/joalberto902/tasks-api/internal/repository"
)

const (
	healthEndpoint = "/health"
	tasksEndpoint  = "/tasks"
	idPath         = "/{id}"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("creating pool: %v", err)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("pinging database: %v", err)
	}

	log.Println("connection to database established")

	repo := repository.NewPostgresRepository(pool)
	taskHandler := handler.NewTaskHandler(repo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+healthEndpoint, handler.HealthHandler)
	mux.HandleFunc("GET "+tasksEndpoint, taskHandler.List)
	mux.HandleFunc("GET "+tasksEndpoint+idPath, taskHandler.Get)
	mux.HandleFunc("POST "+tasksEndpoint, taskHandler.Create)
	mux.HandleFunc("PATCH "+tasksEndpoint+idPath, taskHandler.Update)
	mux.HandleFunc("DELETE "+tasksEndpoint+idPath, taskHandler.Delete)

	server := http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Printf("Serving on port: %s\n", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
