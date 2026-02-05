package config

import (
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Keycloak KeycloakConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port string `env:"SERVER_PORT" envDefault:"3003"`
}

type DatabaseConfig struct {
	Host     string `env:"DB_HOST,required"`
	Port     string `env:"DB_PORT,required"`
	User     string `env:"DB_USER,required"`
	Password string `env:"DB_PASSWORD,required"`
	Name     string `env:"DB_NAME,required"`
}

type KeycloakConfig struct {
	URL           string        `env:"KEYCLOAK_URL,required"`
	Realm         string        `env:"KEYCLOAK_REALM,required"`
	ClientID      string        `env:"KEYCLOAK_CLIENT_ID,required"`
	ClientSecret  string        `env:"KEYCLOAK_CLIENT_SECRET,required"`
	AdminUser     string        `env:"KEYCLOAK_ADMIN_USER,required"`
	AdminPassword string        `env:"KEYCLOAK_ADMIN_PASSWORD,required"`
	AdminRealm    string        `env:"KEYCLOAK_ADMIN_REALM,required"`
	Timeout       time.Duration `env:"KEYCLOAK_TIMEOUT" envDefault:"10s"`
}

type LoggingConfig struct {
	Level string `env:"LOG_LEVEL" envDefault:"error"`
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
