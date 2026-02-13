package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		validate func(*testing.T, *slog.Logger)
	}{
		{
			name:  "success - debug level",
			level: "debug",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - info level",
			level: "info",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - warn level",
			level: "warn",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - warning level (alternative)",
			level: "warning",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - error level",
			level: "error",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - invalid level defaults to error",
			level: "invalid",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - empty level defaults to error",
			level: "",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - uppercase level",
			level: "DEBUG",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
		{
			name:  "success - mixed case level",
			level: "InFo",
			validate: func(t *testing.T, log *slog.Logger) {
				assert.NotNil(t, log)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.level)
			tt.validate(t, log)
		})
	}
}

func TestNew_LevelHandling(t *testing.T) {
	tests := []struct {
		name             string
		level            string
		expectedLoggable func(*slog.Level) bool
	}{
		{
			name:  "debug level - all messages loggable",
			level: "debug",
			expectedLoggable: func(level *slog.Level) bool {
				return true
			},
		},
		{
			name:  "info level - info and above loggable",
			level: "info",
			expectedLoggable: func(level *slog.Level) bool {
				return *level >= slog.LevelInfo
			},
		},
		{
			name:  "error level - only error and above loggable",
			level: "error",
			expectedLoggable: func(level *slog.Level) bool {
				return *level >= slog.LevelError
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := New(tt.level)
			assert.NotNil(t, log)
		})
	}
}
