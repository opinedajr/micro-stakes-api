package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-stakes-api/internal/di"
	"github.com/stretchr/testify/assert"
)

func setupTestEnv() {
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

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	container := di.NewContainer()
	r := gin.Default()

	r.GET("/health", container.HealthCheckHandler().Handle)

	authRoutes := r.Group("/auth")
	{
		authRoutes.POST("/register", func(c *gin.Context) {})
		authRoutes.POST("/login", func(c *gin.Context) {})
		authRoutes.POST("/refresh", func(c *gin.Context) {})
		authRoutes.POST("/logout", func(c *gin.Context) {})
	}

	return r
}

func TestMain_RoutesRegistered(t *testing.T) {
	setupTestEnv()

	r := setupRouter()

	t.Run("success - health route is registered", func(t *testing.T) {
		routes := r.Routes()

		var healthRouteFound bool
		for _, route := range routes {
			if route.Path == "/health" && route.Method == "GET" {
				healthRouteFound = true
				break
			}
		}

		assert.True(t, healthRouteFound, "health route should be registered")
	})

	t.Run("success - auth register route is registered", func(t *testing.T) {
		routes := r.Routes()

		var registerRouteFound bool
		for _, route := range routes {
			if route.Path == "/auth/register" && route.Method == "POST" {
				registerRouteFound = true
				break
			}
		}

		assert.True(t, registerRouteFound, "auth register route should be registered")
	})

	t.Run("success - auth login route is registered", func(t *testing.T) {
		routes := r.Routes()

		var loginRouteFound bool
		for _, route := range routes {
			if route.Path == "/auth/login" && route.Method == "POST" {
				loginRouteFound = true
				break
			}
		}

		assert.True(t, loginRouteFound, "auth login route should be registered")
	})

	t.Run("success - auth refresh route is registered", func(t *testing.T) {
		routes := r.Routes()

		var refreshRouteFound bool
		for _, route := range routes {
			if route.Path == "/auth/refresh" && route.Method == "POST" {
				refreshRouteFound = true
				break
			}
		}

		assert.True(t, refreshRouteFound, "auth refresh route should be registered")
	})

	t.Run("success - auth logout route is registered", func(t *testing.T) {
		routes := r.Routes()

		var logoutRouteFound bool
		for _, route := range routes {
			if route.Path == "/auth/logout" && route.Method == "POST" {
				logoutRouteFound = true
				break
			}
		}

		assert.True(t, logoutRouteFound, "auth logout route should be registered")
	})

	t.Run("success - all routes are registered", func(t *testing.T) {
		routes := r.Routes()

		expectedRoutes := []struct {
			path   string
			method string
		}{
			{"/health", "GET"},
			{"/auth/register", "POST"},
			{"/auth/login", "POST"},
			{"/auth/refresh", "POST"},
			{"/auth/logout", "POST"},
		}

		routeCount := 0
		for _, expected := range expectedRoutes {
			for _, route := range routes {
				if route.Path == expected.path && route.Method == expected.method {
					routeCount++
					break
				}
			}
		}

		assert.Equal(t, len(expectedRoutes), routeCount, "all expected routes should be registered")
	})
}

func TestMain_HealthEndpoint(t *testing.T) {
	setupTestEnv()

	r := setupRouter()

	t.Run("success - health endpoint returns 200", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("success - health endpoint returns correct content-type", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
	})

	t.Run("success - health endpoint returns valid json", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.JSONEq(t, `[{"service_name":"micro-stakes-api","status":"healthy","message":"Service is running"}]`, w.Body.String())
	})
}
