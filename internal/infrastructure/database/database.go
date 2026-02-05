package database

import (
	"fmt"
	"log/slog"

	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresConnection(cfg config.DatabaseConfig, log *slog.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Error("failed to connect to database",
			"host", cfg.Host,
			"port", cfg.Port,
			"database", cfg.Name,
			"error", err)
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := sqlDB.Ping(); err != nil {
		log.Error("failed to ping database",
			"error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("database connection established",
		"host", cfg.Host,
		"database", cfg.Name)

	return db, nil
}
