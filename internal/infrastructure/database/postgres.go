package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresDatabase struct {
	cfg   config.DatabaseConfig
	log   *slog.Logger
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewPostgresDatabase(cfg config.DatabaseConfig, log *slog.Logger) *PostgresDatabase {
	return &PostgresDatabase{
		cfg: cfg,
		log: log,
	}
}

func (p *PostgresDatabase) Connect(ctx context.Context) (*gorm.DB, error) {
	if p.db != nil {
		return p.db, nil
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		p.cfg.Host,
		p.cfg.Port,
		p.cfg.User,
		p.cfg.Password,
		p.cfg.Name,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		p.log.Error("failed to connect to database",
			"host", p.cfg.Host,
			"port", p.cfg.Port,
			"database", p.cfg.Name,
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
		p.log.Error("failed to ping database",
			"error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	p.log.Info("database connection established",
		"host", p.cfg.Host,
		"database", p.cfg.Name)

	p.db = db
	p.sqlDB = sqlDB

	return p.db, nil
}

func (p *PostgresDatabase) Close() error {
	if p.sqlDB != nil {
		return p.sqlDB.Close()
	}
	return nil
}

func (p *PostgresDatabase) Migrate(models ...interface{}) error {
	return nil
}
