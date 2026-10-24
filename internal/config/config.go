package config

import (
	"errors"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

var ErrNoDatabaseURL = errors.New("variable DATABASE_URL is empty, but required")

func Load() (Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db := os.Getenv("DATABASE_URL")
	if db == "" {
		return Config{}, ErrNoDatabaseURL
	}

	return Config{Port: port, DatabaseURL: db}, nil
}
