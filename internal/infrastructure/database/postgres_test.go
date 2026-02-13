package database

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgresDatabase(t *testing.T) {
	tests := []struct {
		name        string
		config      config.DatabaseConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "error - invalid host",
			config: config.DatabaseConfig{
				Host:     "invalid-host-that-does-not-exist-123456789",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
		{
			name: "error - invalid port",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "invalid-port",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
		{
			name: "error - empty host",
			config: config.DatabaseConfig{
				Host:     "",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
			expectError: true,
			errorMsg:    "failed to connect to database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &slog.HandlerOptions{Level: slog.LevelError}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
			ctx := context.Background()
			pgDB := NewPostgresDatabase(tt.config, logger)
			db, err := pgDB.Connect(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestPostgresDatabase_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config config.DatabaseConfig
	}{
		{
			name: "error - connection with valid config structure fails",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &slog.HandlerOptions{Level: slog.LevelError}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
			ctx := context.Background()
			pgDB := NewPostgresDatabase(tt.config, logger)
			db, err := pgDB.Connect(ctx)

			assert.Error(t, err)
			assert.Nil(t, db)

			if err != nil {
				assert.Contains(t, err.Error(), "failed to connect to database")
			}
		})
	}
}

func TestPostgresDatabase_Migrate(t *testing.T) {
	t.Run("success - migrate returns nil (no-op)", func(t *testing.T) {
		opts := &slog.HandlerOptions{Level: slog.LevelError}
		logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		pgDB := NewPostgresDatabase(cfg, logger)

		type TestModel struct {
			ID   uint
			Name string
		}

		err := pgDB.Migrate(&TestModel{})

		assert.NoError(t, err)
	})
}

func TestPostgresDatabase_Close(t *testing.T) {
	t.Run("success - close without connection returns nil", func(t *testing.T) {
		opts := &slog.HandlerOptions{Level: slog.LevelError}
		logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "testuser",
			Password: "testpass",
			Name:     "testdb",
		}

		pgDB := NewPostgresDatabase(cfg, logger)
		err := pgDB.Close()

		assert.NoError(t, err)
	})
}
