package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
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
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "error", err)
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		slog.Error("creating pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		slog.Error("pinging database", "error", err)
		os.Exit(1)
	}

	slog.Info("connection to database established")

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

	slog.Info("serving on port", "port", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
