package healthcheck

import (
	"testing"
)

func TestNewHealthCheckService(t *testing.T) {
	t.Run("success - creates non-nil service", func(t *testing.T) {
		service := NewHealthCheckService()

		if service == nil {
			t.Error("expected service to be non-nil")
		}
	})
}

func TestHealthCheckService_Check(t *testing.T) {
	t.Run("success - returns healthy status", func(t *testing.T) {
		service := NewHealthCheckService()
		result := service.Check()

		if len(result) != 1 {
			t.Errorf("expected 1 health check result, got %d", len(result))
		}

		health := result[0]
		if health.ServiceName != ServiceName {
			t.Errorf("expected service name %s, got %s", ServiceName, health.ServiceName)
		}

		if health.Status != "healthy" {
			t.Errorf("expected status 'healthy', got %s", health.Status)
		}

		if health.Message != "Service is running" {
			t.Errorf("expected message 'Service is running', got %s", health.Message)
		}
	})
}
