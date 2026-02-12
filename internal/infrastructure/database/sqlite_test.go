package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestUser struct {
	ID    uint
	Name  string
	Email string
}

func TestNewSQLiteDatabase(t *testing.T) {
	t.Run("success - creates SQLite database instance", func(t *testing.T) {
		sqliteDB := NewSQLiteDatabase(t)

		assert.NotNil(t, sqliteDB)
		assert.NotNil(t, sqliteDB.t)
		assert.Nil(t, sqliteDB.db)
		assert.Nil(t, sqliteDB.sqlDB)
	})
}

func TestSQLiteDatabase_Connect(t *testing.T) {
	t.Run("success - connects to in-memory database", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		db, err := sqliteDB.Connect(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.NotNil(t, sqliteDB.db)
		assert.NotNil(t, sqliteDB.sqlDB)
	})

	t.Run("success - connect is idempotent (returns same instance)", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		db1, err1 := sqliteDB.Connect(ctx)
		assert.NoError(t, err1)

		db2, err2 := sqliteDB.Connect(ctx)
		assert.NoError(t, err2)

		assert.Equal(t, db1, db2)
	})
}

func TestSQLiteDatabase_Migrate(t *testing.T) {
	t.Run("success - migrates single model", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		_, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)

		err = sqliteDB.Migrate(&TestUser{})

		assert.NoError(t, err)

		var tableName string
		result := sqliteDB.db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='test_users'").Scan(&tableName)
		assert.NoError(t, result.Error)
		assert.Equal(t, "test_users", tableName)
	})

	t.Run("success - migrates multiple models", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		_, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)

		type TestProduct struct {
			ID    uint
			Name  string
			Price float64
		}

		err = sqliteDB.Migrate(&TestUser{}, &TestProduct{})

		assert.NoError(t, err)
	})

	t.Run("error - migrate without connection", func(t *testing.T) {
		sqliteDB := NewSQLiteDatabase(t)

		err := sqliteDB.Migrate(&TestUser{})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database not connected")
	})

	t.Run("error - migrate with no models", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		_, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)

		err = sqliteDB.Migrate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no models provided for migration")
	})
}

func TestSQLiteDatabase_Close(t *testing.T) {
	t.Run("success - closes connection", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		_, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)

		err = sqliteDB.Close()

		assert.NoError(t, err)
	})

	t.Run("success - close without connection returns nil", func(t *testing.T) {
		sqliteDB := NewSQLiteDatabase(t)

		err := sqliteDB.Close()

		assert.NoError(t, err)
	})
}

func TestSQLiteDatabase_IntegrationWorkflow(t *testing.T) {
	t.Run("success - full workflow (connect -> migrate -> query -> close)", func(t *testing.T) {
		ctx := context.Background()
		sqliteDB := NewSQLiteDatabase(t)

		db, err := sqliteDB.Connect(ctx)
		assert.NoError(t, err)

		err = sqliteDB.Migrate(&TestUser{})
		assert.NoError(t, err)

		testUser := TestUser{Name: "John Doe", Email: "john@example.com"}
		result := db.Create(&testUser)
		assert.NoError(t, result.Error)
		assert.NotZero(t, testUser.ID)

		var retrievedUser TestUser
		result = db.First(&retrievedUser, testUser.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, "John Doe", retrievedUser.Name)
		assert.Equal(t, "john@example.com", retrievedUser.Email)

		err = sqliteDB.Close()
		assert.NoError(t, err)
	})
}
