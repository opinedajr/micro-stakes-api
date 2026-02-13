package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Load_Success(t *testing.T) {
	tests := []struct {
		name     string
		setEnv   func() func()
		validate func(*testing.T, *Config, error)
	}{
		{
			name: "success - load all required env vars",
			setEnv: func() func() {
				os.Setenv("SERVER_PORT", "8080")
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "testpass")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("KEYCLOAK_URL", "http://keycloak:8080")
				os.Setenv("KEYCLOAK_REALM", "test-realm")
				os.Setenv("KEYCLOAK_CLIENT_ID", "test-client")
				os.Setenv("KEYCLOAK_CLIENT_SECRET", "test-secret")
				os.Setenv("KEYCLOAK_ADMIN_USER", "admin")
				os.Setenv("KEYCLOAK_ADMIN_PASSWORD", "admin-pass")
				os.Setenv("KEYCLOAK_ADMIN_REALM", "master")
				os.Setenv("LOG_LEVEL", "debug")
				return func() {
					os.Unsetenv("SERVER_PORT")
					os.Unsetenv("DB_HOST")
					os.Unsetenv("DB_PORT")
					os.Unsetenv("DB_USER")
					os.Unsetenv("DB_PASSWORD")
					os.Unsetenv("DB_NAME")
					os.Unsetenv("KEYCLOAK_URL")
					os.Unsetenv("KEYCLOAK_REALM")
					os.Unsetenv("KEYCLOAK_CLIENT_ID")
					os.Unsetenv("KEYCLOAK_CLIENT_SECRET")
					os.Unsetenv("KEYCLOAK_ADMIN_USER")
					os.Unsetenv("KEYCLOAK_ADMIN_PASSWORD")
					os.Unsetenv("KEYCLOAK_ADMIN_REALM")
					os.Unsetenv("LOG_LEVEL")
				}
			},
			validate: func(t *testing.T, cfg *Config, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, "8080", cfg.Server.Port)
				assert.Equal(t, "localhost", cfg.Database.Host)
				assert.Equal(t, "5432", cfg.Database.Port)
				assert.Equal(t, "testuser", cfg.Database.User)
				assert.Equal(t, "testpass", cfg.Database.Password)
				assert.Equal(t, "testdb", cfg.Database.Name)
				assert.Equal(t, "http://keycloak:8080", cfg.Keycloak.URL)
				assert.Equal(t, "test-realm", cfg.Keycloak.Realm)
				assert.Equal(t, "test-client", cfg.Keycloak.ClientID)
				assert.Equal(t, "test-secret", cfg.Keycloak.ClientSecret)
				assert.Equal(t, "admin", cfg.Keycloak.AdminUser)
				assert.Equal(t, "admin-pass", cfg.Keycloak.AdminPassword)
				assert.Equal(t, "master", cfg.Keycloak.AdminRealm)
				assert.Equal(t, "debug", cfg.Logging.Level)
			},
		},
		{
			name: "success - load with default values",
			setEnv: func() func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "testpass")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("KEYCLOAK_URL", "http://keycloak:8080")
				os.Setenv("KEYCLOAK_REALM", "test-realm")
				os.Setenv("KEYCLOAK_CLIENT_ID", "test-client")
				os.Setenv("KEYCLOAK_CLIENT_SECRET", "test-secret")
				os.Setenv("KEYCLOAK_ADMIN_USER", "admin")
				os.Setenv("KEYCLOAK_ADMIN_PASSWORD", "admin-pass")
				os.Setenv("KEYCLOAK_ADMIN_REALM", "master")
				return func() {
					os.Unsetenv("SERVER_PORT")
					os.Unsetenv("DB_HOST")
					os.Unsetenv("DB_PORT")
					os.Unsetenv("DB_USER")
					os.Unsetenv("DB_PASSWORD")
					os.Unsetenv("DB_NAME")
					os.Unsetenv("KEYCLOAK_URL")
					os.Unsetenv("KEYCLOAK_REALM")
					os.Unsetenv("KEYCLOAK_CLIENT_ID")
					os.Unsetenv("KEYCLOAK_CLIENT_SECRET")
					os.Unsetenv("KEYCLOAK_ADMIN_USER")
					os.Unsetenv("KEYCLOAK_ADMIN_PASSWORD")
					os.Unsetenv("KEYCLOAK_ADMIN_REALM")
				}
			},
			validate: func(t *testing.T, cfg *Config, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				assert.Equal(t, "3003", cfg.Server.Port)
				assert.Equal(t, "error", cfg.Logging.Level)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setEnv()
			defer cleanup()

			cfg, err := Load()
			tt.validate(t, cfg, err)
		})
	}
}

func TestConfig_Load_MissingRequiredEnv(t *testing.T) {
	tests := []struct {
		name     string
		unsetEnv []string
		setEnv   func() func()
	}{
		{
			name:     "error - missing DB_HOST",
			unsetEnv: []string{"DB_HOST"},
			setEnv: func() func() {
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "testpass")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("KEYCLOAK_URL", "http://keycloak:8080")
				os.Setenv("KEYCLOAK_REALM", "test-realm")
				os.Setenv("KEYCLOAK_CLIENT_ID", "test-client")
				os.Setenv("KEYCLOAK_CLIENT_SECRET", "test-secret")
				os.Setenv("KEYCLOAK_ADMIN_USER", "admin")
				os.Setenv("KEYCLOAK_ADMIN_PASSWORD", "admin-pass")
				os.Setenv("KEYCLOAK_ADMIN_REALM", "master")
				return func() {
					os.Unsetenv("DB_PORT")
					os.Unsetenv("DB_USER")
					os.Unsetenv("DB_PASSWORD")
					os.Unsetenv("DB_NAME")
					os.Unsetenv("KEYCLOAK_URL")
					os.Unsetenv("KEYCLOAK_REALM")
					os.Unsetenv("KEYCLOAK_CLIENT_ID")
					os.Unsetenv("KEYCLOAK_CLIENT_SECRET")
					os.Unsetenv("KEYCLOAK_ADMIN_USER")
					os.Unsetenv("KEYCLOAK_ADMIN_PASSWORD")
					os.Unsetenv("KEYCLOAK_ADMIN_REALM")
				}
			},
		},
		{
			name:     "error - missing KEYCLOAK_URL",
			unsetEnv: []string{"KEYCLOAK_URL"},
			setEnv: func() func() {
				os.Setenv("DB_HOST", "localhost")
				os.Setenv("DB_PORT", "5432")
				os.Setenv("DB_USER", "testuser")
				os.Setenv("DB_PASSWORD", "testpass")
				os.Setenv("DB_NAME", "testdb")
				os.Setenv("KEYCLOAK_REALM", "test-realm")
				os.Setenv("KEYCLOAK_CLIENT_ID", "test-client")
				os.Setenv("KEYCLOAK_CLIENT_SECRET", "test-secret")
				os.Setenv("KEYCLOAK_ADMIN_USER", "admin")
				os.Setenv("KEYCLOAK_ADMIN_PASSWORD", "admin-pass")
				os.Setenv("KEYCLOAK_ADMIN_REALM", "master")
				return func() {
					os.Unsetenv("DB_HOST")
					os.Unsetenv("DB_PORT")
					os.Unsetenv("DB_USER")
					os.Unsetenv("DB_PASSWORD")
					os.Unsetenv("DB_NAME")
					os.Unsetenv("KEYCLOAK_REALM")
					os.Unsetenv("KEYCLOAK_CLIENT_ID")
					os.Unsetenv("KEYCLOAK_CLIENT_SECRET")
					os.Unsetenv("KEYCLOAK_ADMIN_USER")
					os.Unsetenv("KEYCLOAK_ADMIN_PASSWORD")
					os.Unsetenv("KEYCLOAK_ADMIN_REALM")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, env := range tt.unsetEnv {
				os.Unsetenv(env)
			}
			cleanup := tt.setEnv()
			defer cleanup()

			cfg, err := Load()
			assert.Error(t, err)
			assert.Nil(t, cfg)
		})
	}
}
