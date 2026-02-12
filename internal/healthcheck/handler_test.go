package healthcheck

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type MockService struct{}

func (m *MockService) Check() []Health {
	return []Health{
		{
			ServiceName: ServiceName,
			Status:      "healthy",
			Message:     "Service is running",
		},
	}
}

func TestNewHandler(t *testing.T) {
	t.Run("success - creates handler with service", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		if handler == nil {
			t.Error("expected handler to be non-nil")
		}

		if handler.service == nil {
			t.Error("expected service to be non-nil")
		}
	})
}

func TestHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - returns 200 status code", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		handler.Handle(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status code 200, got %d", w.Code)
		}
	})

	t.Run("success - returns correct JSON structure", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		handler.Handle(c)

		if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
			t.Errorf("expected content-type application/json, got %s", w.Header().Get("Content-Type"))
		}

		var response []Health
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(response) != 1 {
			t.Fatalf("expected 1 health check result, got %d", len(response))
		}

		health := response[0]
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

	t.Run("success - returns valid JSON", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/health", nil)

		handler.Handle(c)

		var buf bytes.Buffer
		if err := json.Compact(&buf, w.Body.Bytes()); err != nil {
			t.Errorf("response is not valid JSON: %v", err)
		}

		var response interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("failed to unmarshal JSON: %v", err)
		}
	})
}
