package identity

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
	"github.com/stretchr/testify/assert"
)

func TestNewKeycloakAdapter(t *testing.T) {
	tests := []struct {
		name        string
		config      config.KeycloakConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "error - invalid Keycloak URL",
			config: config.KeycloakConfig{
				URL:           "http://invalid-keycloak-that-does-not-exist-123456789:9999",
				Realm:         "test-realm",
				ClientID:      "test-client",
				ClientSecret:  "test-secret",
				AdminUser:     "admin",
				AdminPassword: "admin-password",
				AdminRealm:    "master",
				Timeout:       10 * time.Second,
			},
			expectError: true,
			errorMsg:    "failed to obtain admin token",
		},
		{
			name: "error - empty URL",
			config: config.KeycloakConfig{
				URL:           "",
				Realm:         "test-realm",
				ClientID:      "test-client",
				ClientSecret:  "test-secret",
				AdminUser:     "admin",
				AdminPassword: "admin-password",
				AdminRealm:    "master",
				Timeout:       10 * time.Second,
			},
			expectError: true,
			errorMsg:    "failed to obtain admin token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &slog.HandlerOptions{Level: slog.LevelError}
			logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

			adapter, err := NewKeycloakAdapter(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, adapter)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, adapter)
			}
		})
	}
}

func TestKeycloakAdapter_ensureAdminToken(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*testing.T) (*KeycloakAdapter, context.Context)
		expectError bool
	}{
		{
			name: "error - admin token not initialized",
			setup: func(t *testing.T) (*KeycloakAdapter, context.Context) {
				opts := &slog.HandlerOptions{Level: slog.LevelError}
				adapter := &KeycloakAdapter{
					client:       gocloak.NewClient("http://localhost:8080"),
					config:       config.KeycloakConfig{},
					logger:       slog.New(slog.NewJSONHandler(os.Stdout, opts)),
					adminToken:   nil,
					tokenExpires: time.Time{},
				}
				return adapter, context.Background()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, ctx := tt.setup(t)

			err := adapter.ensureAdminToken(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeycloakAdapter_retryWithBackoff(t *testing.T) {
	opts := &slog.HandlerOptions{Level: slog.LevelError}
	adapter := &KeycloakAdapter{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, opts)),
	}

	tests := []struct {
		name        string
		operation   func() error
		expectError bool
	}{
		{
			name: "success - operation succeeds on first try",
			operation: func() error {
				return nil
			},
			expectError: false,
		},
		{
			name: "error - operation fails permanently",
			operation: func() error {
				return errors.New("permanent failure")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.retryWithBackoff(tt.operation)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
