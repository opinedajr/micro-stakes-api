package bankroll

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db.AutoMigrate(&Bankroll{})

	return db
}

func TestPostgresBankrollRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll)

		assert.NoError(t, err)
		assert.NotZero(t, bankroll.ID)
		assert.Equal(t, uint(1), bankroll.UserID)
		assert.Equal(t, "Main Bankroll", bankroll.Name)
		assert.Equal(t, CurrencyBRL, bankroll.Currency)
		assert.Equal(t, 1000.00, bankroll.InitialBalance)
		assert.Equal(t, 1000.00, bankroll.CurrentBalance)
	})

	t.Run("duplicate name per user", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll1 := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll1)
		require.NoError(t, err)

		bankroll2 := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyUSD,
			InitialBalance:       500.00,
			CurrentBalance:       500.00,
			StartDate:            startDate,
			CommissionPercentage: 3.0,
		}

		err = repo.Create(ctx, bankroll2)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBankrollNameExists)
	})

	t.Run("same name different user", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll1 := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll1)
		require.NoError(t, err)

		bankroll2 := &Bankroll{
			UserID:               2,
			Name:                 "Main Bankroll",
			Currency:             CurrencyUSD,
			InitialBalance:       500.00,
			CurrentBalance:       500.00,
			StartDate:            startDate,
			CommissionPercentage: 3.0,
		}

		err = repo.Create(ctx, bankroll2)

		assert.NoError(t, err)
		assert.NotZero(t, bankroll2.ID)
	})
}

func TestPostgresBankrollRepository_ListByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll1 := &Bankroll{
			UserID:               1,
			Name:                 "Bankroll 1",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		bankroll2 := &Bankroll{
			UserID:               1,
			Name:                 "Bankroll 2",
			Currency:             CurrencyUSD,
			InitialBalance:       500.00,
			CurrentBalance:       500.00,
			StartDate:            startDate,
			CommissionPercentage: 3.0,
		}

		bankroll3 := &Bankroll{
			UserID:               2,
			Name:                 "Other User Bankroll",
			Currency:             CurrencyEUR,
			InitialBalance:       2000.00,
			CurrentBalance:       2000.00,
			StartDate:            startDate,
			CommissionPercentage: 7.0,
		}

		err = repo.Create(ctx, bankroll1)
		require.NoError(t, err)

		err = repo.Create(ctx, bankroll2)
		require.NoError(t, err)

		err = repo.Create(ctx, bankroll3)
		require.NoError(t, err)

		bankrolls, err := repo.ListByUserID(ctx, 1)

		assert.NoError(t, err)
		assert.Len(t, bankrolls, 2)

		bankrollIDs := make([]uint, len(bankrolls))
		for i, b := range bankrolls {
			bankrollIDs[i] = b.ID
		}
		assert.Contains(t, bankrollIDs, bankroll1.ID)
		assert.Contains(t, bankrollIDs, bankroll2.ID)
	})

	t.Run("empty list", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		bankrolls, err := repo.ListByUserID(ctx, 999)

		assert.NoError(t, err)
		assert.Empty(t, bankrolls)
	})
}

func TestPostgresBankrollRepository_FindByID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, bankroll.ID, 1)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, bankroll.ID, found.ID)
		assert.Equal(t, "Main Bankroll", found.Name)
	})

	t.Run("not found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		found, err := repo.FindByID(ctx, 999, 1)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		assert.Nil(t, found)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, bankroll.ID, 2)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		assert.Nil(t, found)
	})
}

func TestPostgresBankrollRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll)
		require.NoError(t, err)

		bankroll.Name = "Updated Bankroll"
		bankroll.CommissionPercentage = 3.0

		err = repo.Update(ctx, bankroll)

		assert.NoError(t, err)

		updated, err := repo.FindByID(ctx, bankroll.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, "Updated Bankroll", updated.Name)
		assert.Equal(t, 3.0, updated.CommissionPercentage)
		assert.Equal(t, 1000.00, updated.InitialBalance)
		assert.Equal(t, 1000.00, updated.CurrentBalance)
	})

	t.Run("duplicate name per user", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll1 := &Bankroll{
			UserID:               1,
			Name:                 "Bankroll 1",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		bankroll2 := &Bankroll{
			UserID:               1,
			Name:                 "Bankroll 2",
			Currency:             CurrencyUSD,
			InitialBalance:       500.00,
			CurrentBalance:       500.00,
			StartDate:            startDate,
			CommissionPercentage: 3.0,
		}

		err = repo.Create(ctx, bankroll1)
		require.NoError(t, err)

		err = repo.Create(ctx, bankroll2)
		require.NoError(t, err)

		bankroll1.Name = "Bankroll 2"
		err = repo.Update(ctx, bankroll1)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBankrollNameExists)
	})

	t.Run("same name different user", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll1 := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		bankroll2 := &Bankroll{
			UserID:               2,
			Name:                 "Main Bankroll",
			Currency:             CurrencyUSD,
			InitialBalance:       500.00,
			CurrentBalance:       500.00,
			StartDate:            startDate,
			CommissionPercentage: 3.0,
		}

		err = repo.Create(ctx, bankroll1)
		require.NoError(t, err)

		err = repo.Create(ctx, bankroll2)
		require.NoError(t, err)

		bankroll2.CommissionPercentage = 4.0
		err = repo.Update(ctx, bankroll2)

		assert.NoError(t, err)

		updated, err := repo.FindByID(ctx, bankroll2.ID, 2)
		require.NoError(t, err)
		assert.Equal(t, 4.0, updated.CommissionPercentage)
	})
}

func TestPostgresBankrollRepository_Reset(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		startDate, err := time.Parse("2006-01-02", "2026-02-01")
		require.NoError(t, err)

		bankroll := &Bankroll{
			UserID:               1,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			StartDate:            startDate,
			CommissionPercentage: 5.0,
		}

		err = repo.Create(ctx, bankroll)
		require.NoError(t, err)

		err = repo.Reset(ctx, bankroll.ID, 1)

		assert.NoError(t, err)

		reset, err := repo.FindByID(ctx, bankroll.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, 0.0, reset.InitialBalance)
		assert.Equal(t, 0.0, reset.CurrentBalance)
	})

	t.Run("not found", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresBankrollRepository(db)
		ctx := context.Background()

		err := repo.Reset(ctx, 999, 1)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
	})
}
