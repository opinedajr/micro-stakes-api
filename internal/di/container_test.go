package di

import (
	"os"
	"testing"
)

func setupEnvVars() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "test")
	os.Setenv("DB_PASSWORD", "test")
	os.Setenv("DB_NAME", "test")
	os.Setenv("KEYCLOAK_URL", "http://localhost:8080")
	os.Setenv("KEYCLOAK_CLIENT_ID", "test")
	os.Setenv("KEYCLOAK_CLIENT_SECRET", "test")
	os.Setenv("KEYCLOAK_REALM", "test")
	os.Setenv("KEYCLOAK_ADMIN_USER", "admin")
	os.Setenv("KEYCLOAK_ADMIN_PASSWORD", "admin")
	os.Setenv("KEYCLOAK_ADMIN_REALM", "master")
	os.Setenv("LOG_LEVEL", "error")
}

func TestNewContainer(t *testing.T) {
	setupEnvVars()

	t.Run("success - creates container with nil dependencies", func(t *testing.T) {
		container := NewContainer()

		if container == nil {
			t.Fatal("expected container to be non-nil")
		}

		if container.repositories == nil {
			t.Error("expected repositories to be initialized")
		}

		if container.services == nil {
			t.Error("expected services to be initialized")
		}

		if container.handlers == nil {
			t.Error("expected handlers to be initialized")
		}

		if container.middlewares == nil {
			t.Error("expected middlewares to be initialized")
		}

		if container.config != nil {
			t.Error("expected config to be nil initially")
		}

		if container.logger != nil {
			t.Error("expected logger to be nil initially")
		}

		if container.db != nil {
			t.Error("expected db to be nil initially")
		}

		if container.identityClient != nil {
			t.Error("expected identityClient to be nil initially")
		}
	})
}

func TestContainer_HealthCheckService(t *testing.T) {
	setupEnvVars()

	t.Run("success - creates service on first call", func(t *testing.T) {
		container := NewContainer()

		service := container.HealthCheckService()

		if service == nil {
			t.Error("expected service to be non-nil")
		}
	})

	t.Run("success - returns same instance on subsequent calls", func(t *testing.T) {
		container := NewContainer()

		service1 := container.HealthCheckService()
		service2 := container.HealthCheckService()

		if service1 != service2 {
			t.Error("expected same instance on subsequent calls")
		}
	})
}

func TestContainer_HealthCheckHandler(t *testing.T) {
	setupEnvVars()

	t.Run("success - creates handler on first call", func(t *testing.T) {
		container := NewContainer()

		handler := container.HealthCheckHandler()

		if handler == nil {
			t.Error("expected handler to be non-nil")
		}
	})

	t.Run("success - returns same instance on subsequent calls", func(t *testing.T) {
		container := NewContainer()

		handler1 := container.HealthCheckHandler()
		handler2 := container.HealthCheckHandler()

		if handler1 != handler2 {
			t.Error("expected same instance on subsequent calls")
		}
	})
}

func TestContainer_Config(t *testing.T) {
	setupEnvVars()

	t.Run("success - creates config on first call", func(t *testing.T) {
		container := NewContainer()

		config := container.Config()

		if config == nil {
			t.Error("expected config to be non-nil")
		}
	})

	t.Run("success - returns same instance on subsequent calls", func(t *testing.T) {
		container := NewContainer()

		config1 := container.Config()
		config2 := container.Config()

		if config1 != config2 {
			t.Error("expected same instance on subsequent calls")
		}
	})
}

func TestContainer_Logger(t *testing.T) {
	setupEnvVars()

	t.Run("success - creates logger on first call", func(t *testing.T) {
		container := NewContainer()

		logger := container.Logger()

		if logger == nil {
			t.Error("expected logger to be non-nil")
		}
	})

	t.Run("success - returns same instance on subsequent calls", func(t *testing.T) {
		container := NewContainer()

		logger1 := container.Logger()
		logger2 := container.Logger()

		if logger1 != logger2 {
			t.Error("expected same instance on subsequent calls")
		}
	})
}
