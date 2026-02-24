package di

import (
	"context"
	"log/slog"

	"github.com/opinedajr/micro-stakes-api/internal/auth"
	"github.com/opinedajr/micro-stakes-api/internal/bankroll"
	"github.com/opinedajr/micro-stakes-api/internal/healthcheck"
	"github.com/opinedajr/micro-stakes-api/internal/infrastructure/database"
	"github.com/opinedajr/micro-stakes-api/internal/infrastructure/identity"
	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
	"github.com/opinedajr/micro-stakes-api/internal/shared/logger"
	"gorm.io/gorm"
)

type Container struct {
	config         *config.Config
	logger         *slog.Logger
	db             *gorm.DB
	identityClient identity.IdentityProvider
	repositories   *RepositoryDependencies
	services       *ServiceDependencies
	handlers       *HandlerDependencies
}

type RepositoryDependencies struct {
	userRepository     auth.UserRepository
	bankrollRepository bankroll.BankrollRepository
}

type HandlerDependencies struct {
	healthcheckHandler *healthcheck.Handler
	authHandler        *auth.AuthHandler
	bankrollHandler    *bankroll.BankrollHandler
}

type ServiceDependencies struct {
	healthcheckService *healthcheck.Service
	authService        auth.AuthService
	bankrollService    bankroll.BankrollService
}

func NewContainer() *Container {
	return &Container{
		repositories: &RepositoryDependencies{},
		services:     &ServiceDependencies{},
		handlers:     &HandlerDependencies{},
	}
}

func (c *Container) Config() *config.Config {
	if c.config == nil {
		cfg, err := config.Load()
		if err != nil {
			panic("failed to load config: " + err.Error())
		}
		c.config = cfg
	}
	return c.config
}

func (c *Container) Logger() *slog.Logger {
	if c.logger == nil {
		c.logger = logger.New(c.Config().Logging.Level)
	}
	return c.logger
}

func (c *Container) DB() *gorm.DB {
	if c.db == nil {
		ctx := context.Background()
		pgDB := database.NewPostgresDatabase(c.Config().Database, c.Logger())
		db, err := pgDB.Connect(ctx)
		if err != nil {
			panic("failed to connect to database: " + err.Error())
		}
		c.db = db
	}
	return c.db
}

func (c *Container) IdentityProvider() identity.IdentityProvider {
	if c.identityClient == nil {
		provider, err := identity.NewKeycloakAdapter(c.Config().Keycloak, c.Logger())
		if err != nil {
			panic("failed to create identity provider: " + err.Error())
		}
		c.identityClient = provider
	}
	return c.identityClient
}

func (c *Container) HealthCheckService() *healthcheck.Service {
	if c.services.healthcheckService == nil {
		c.services.healthcheckService = healthcheck.NewHealthCheckService()
	}
	return c.services.healthcheckService
}

func (c *Container) HealthCheckHandler() *healthcheck.Handler {
	if c.handlers.healthcheckHandler == nil {
		c.handlers.healthcheckHandler = healthcheck.NewHandler(c.HealthCheckService())
	}
	return c.handlers.healthcheckHandler
}

func (c *Container) UserRepository() auth.UserRepository {
	if c.repositories.userRepository == nil {
		c.repositories.userRepository = auth.NewPostgresUserRepository(c.DB())
	}
	return c.repositories.userRepository
}

func (c *Container) AuthService() auth.AuthService {
	if c.services.authService == nil {
		c.services.authService = auth.NewAuthService(
			c.UserRepository(),
			c.IdentityProvider(),
			c.Logger(),
		)
	}
	return c.services.authService
}

func (c *Container) AuthHandler() *auth.AuthHandler {
	if c.handlers.authHandler == nil {
		c.handlers.authHandler = auth.NewAuthHandler(
			c.AuthService(),
			c.Logger(),
		)
	}
	return c.handlers.authHandler
}

func (c *Container) BankrollRepository() bankroll.BankrollRepository {
	if c.repositories.bankrollRepository == nil {
		c.repositories.bankrollRepository = bankroll.NewPostgresBankrollRepository(c.DB())
	}
	return c.repositories.bankrollRepository
}

func (c *Container) BankrollService() bankroll.BankrollService {
	if c.services.bankrollService == nil {
		c.services.bankrollService = bankroll.NewBankrollService(
			c.BankrollRepository(),
			c.Logger(),
		)
	}
	return c.services.bankrollService
}

func (c *Container) BankrollHandler() *bankroll.BankrollHandler {
	if c.handlers.bankrollHandler == nil {
		c.handlers.bankrollHandler = bankroll.NewBankrollHandler(
			c.BankrollService(),
			c.Logger(),
		)
	}
	return c.handlers.bankrollHandler
}
