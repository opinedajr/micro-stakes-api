package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLiteDatabase struct {
	t     *testing.T
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewSQLiteDatabase(t *testing.T) *SQLiteDatabase {
	return &SQLiteDatabase{
		t: t,
	}
}

func (s *SQLiteDatabase) Connect(ctx context.Context) (*gorm.DB, error) {
	if s.db != nil {
		return s.db, nil
	}

	s.t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create test database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	s.t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			s.t.Logf("warning: failed to close database connection: %v", err)
		}
	})

	s.db = db
	s.sqlDB = sqlDB

	return s.db, nil
}

func (s *SQLiteDatabase) Close() error {
	if s.sqlDB != nil {
		return s.sqlDB.Close()
	}
	return nil
}

func (s *SQLiteDatabase) Migrate(models ...interface{}) error {
	if len(models) == 0 {
		return fmt.Errorf("no models provided for migration")
	}
	if s.db == nil {
		return fmt.Errorf("database not connected")
	}
	return s.db.AutoMigrate(models...)
}
