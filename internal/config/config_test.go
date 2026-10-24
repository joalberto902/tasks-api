package config

import (
	"errors"
	"testing"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		port           string
		expectedConfig Config
		expectedErr    error
	}{
		{
			name: "DATABASE_URL and empty port",
			url:  "postgres://postgres:dev@localhost:5432/tasks",
			expectedConfig: Config{
				DatabaseURL: "postgres://postgres:dev@localhost:5432/tasks",
				Port:        "8080",
			},
			expectedErr: nil,
		},
		{
			name: "DATABASE_URL and PORT",
			url:  "postgres://postgres:dev@localhost:5432/tasks",
			port: "9090",
			expectedConfig: Config{
				DatabaseURL: "postgres://postgres:dev@localhost:5432/tasks",
				Port:        "9090",
			},
			expectedErr: nil,
		},
		{
			name:           "empty DATABASE_URL",
			expectedConfig: Config{},
			expectedErr:    ErrNoDatabaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", tt.url)
			t.Setenv("PORT", tt.port)

			cfg, err := Load()

			if cfg != tt.expectedConfig {
				t.Fatalf("wrong Config struct, expected=%+v, got=%+v",
					tt.expectedConfig, cfg)
			}

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("wrong error, expected=%v, got=%v",
					tt.expectedErr, err)
			}
		})
	}
}
