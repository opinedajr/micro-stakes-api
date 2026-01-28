package di

import (
	"github.com/opinedajr/micro-stakes-api/internal/healthcheck"
)

type Container struct {
	repositories *RepositoryDependencies
	services     *ServiceDependencies
	handlers     *HandlerDependencies
	middlewares  *middlewareDependencies
}

type RepositoryDependencies struct {
}

type HandlerDependencies struct {
	healthcheckHandler *healthcheck.Handler
}

type ServiceDependencies struct {
	healthcheckService *healthcheck.Service
}

type middlewareDependencies struct{}

func NewContainer() *Container {
	return &Container{
		repositories: &RepositoryDependencies{},
		services:     &ServiceDependencies{},
		handlers:     &HandlerDependencies{},
		middlewares:  &middlewareDependencies{},
	}
}

// Services
func (c *Container) HealthCheckService() *healthcheck.Service {
	if c.services.healthcheckService == nil {
		c.services.healthcheckService = healthcheck.NewHealthCheckService()
	}
	return c.services.healthcheckService
}

// Handlers
func (c *Container) HealthCheckHandler() *healthcheck.Handler {
	if c.handlers.healthcheckHandler == nil {
		c.handlers.healthcheckHandler = healthcheck.NewHandler(c.HealthCheckService())
	}
	return c.handlers.healthcheckHandler
}
