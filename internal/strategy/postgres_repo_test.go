package strategy

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

	db.AutoMigrate(&Strategy{})

	return db
}

func TestPostgresStrategyRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Value Back Strategy",
			Description:  stringPtr("Back high-value selections at odds > 2.0"),
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}

		err := repo.Create(ctx, strategy)

		assert.NoError(t, err)
		assert.NotZero(t, strategy.ID)
		assert.Equal(t, uint(1), strategy.UserID)
		assert.Equal(t, "Value Back Strategy", strategy.Name)
		assert.Equal(t, "Back high-value selections at odds > 2.0", *strategy.Description)
		assert.Equal(t, 10.00, strategy.DefaultStake)
		assert.Equal(t, StrategyTypeBack, strategy.Type)
		assert.True(t, strategy.Active)
	})

	t.Run("success with nil description", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Lay the Draw",
			Description:  nil,
			DefaultStake: 5.50,
			Type:         StrategyTypeLay,
			Active:       false,
		}

		err := repo.Create(ctx, strategy)

		assert.NoError(t, err)
		assert.NotZero(t, strategy.ID)
		assert.Nil(t, strategy.Description)
	})
}

func TestPostgresStrategyRepository_ListByUserID(t *testing.T) {
	t.Run("success with multiple rows", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy1 := &Strategy{
			UserID:       1,
			Name:         "Strategy 1",
			Description:  stringPtr("First strategy"),
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		strategy2 := &Strategy{
			UserID:       1,
			Name:         "Strategy 2",
			Description:  stringPtr("Second strategy"),
			DefaultStake: 20.00,
			Type:         StrategyTypeLay,
			Active:       false,
		}
		require.NoError(t, repo.Create(ctx, strategy1))
		require.NoError(t, repo.Create(ctx, strategy2))

		strategies, total, err := repo.ListByUserID(ctx, 1, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, strategies, 2)
		assert.Equal(t, "Strategy 2", strategies[0].Name)
		assert.Equal(t, "Strategy 1", strategies[1].Name)
	})

	t.Run("empty result when user has no strategies", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategies, total, err := repo.ListByUserID(ctx, 1, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Len(t, strategies, 0)
	})

	t.Run("out-of-bounds page returns empty slice with correct total", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		strategies, total, err := repo.ListByUserID(ctx, 1, 9999, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, strategies, 0)
	})

	t.Run("strategies from other users excluded", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy1 := &Strategy{
			UserID:       1,
			Name:         "User 1 Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		strategy2 := &Strategy{
			UserID:       2,
			Name:         "User 2 Strategy",
			DefaultStake: 20.00,
			Type:         StrategyTypeLay,
			Active:       false,
		}
		require.NoError(t, repo.Create(ctx, strategy1))
		require.NoError(t, repo.Create(ctx, strategy2))

		strategies, total, err := repo.ListByUserID(ctx, 1, 1, 20)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, strategies, 1)
		assert.Equal(t, "User 1 Strategy", strategies[0].Name)
	})
}

func TestPostgresStrategyRepository_FindByID(t *testing.T) {
	t.Run("success returns correct strategy for owner", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Value Back Strategy",
			Description:  stringPtr("Back high-value selections at odds > 2.0"),
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		found, err := repo.FindByID(ctx, strategy.ID, 1)

		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, strategy.ID, found.ID)
		assert.Equal(t, uint(1), found.UserID)
		assert.Equal(t, "Value Back Strategy", found.Name)
		assert.Equal(t, "Back high-value selections at odds > 2.0", *found.Description)
		assert.Equal(t, 10.00, found.DefaultStake)
		assert.Equal(t, StrategyTypeBack, found.Type)
		assert.True(t, found.Active)
	})

	t.Run("not found returns ErrStrategyNotFound", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		found, err := repo.FindByID(ctx, 99999, 1)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
	})

	t.Run("foreign user ID returns ErrStrategyNotFound", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "User 1 Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		found, err := repo.FindByID(ctx, strategy.ID, 2)

		assert.Error(t, err)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
	})
}

func TestPostgresStrategyRepository_Update(t *testing.T) {
	t.Run("success all fields updated updated_at refreshed", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Original Name",
			Description:  stringPtr("Original description"),
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		originalUpdatedAt := strategy.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		strategy.Name = "Updated Name"
		updatedDesc := "Updated description"
		strategy.Description = &updatedDesc
		strategy.DefaultStake = 15.00
		strategy.Type = StrategyTypeLay
		strategy.Active = false

		err := repo.Update(ctx, strategy)

		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, strategy.ID, 1)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", found.Name)
		assert.Equal(t, "Updated description", *found.Description)
		assert.Equal(t, 15.00, found.DefaultStake)
		assert.Equal(t, StrategyTypeLay, found.Type)
		assert.False(t, found.Active)
		assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("not found returns ErrStrategyNotFound", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			ID:           99999,
			UserID:       1,
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}

		err := repo.Update(ctx, strategy)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
	})

	t.Run("foreign user ID returns ErrStrategyNotFound", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "User 1 Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		strategy.UserID = 2
		strategy.Name = "Updated Name"

		err := repo.Update(ctx, strategy)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
	})
}

func TestPostgresStrategyRepository_UpdateStatus(t *testing.T) {
	t.Run("activate sets active=true and updates updated_at", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       false,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		originalUpdatedAt := strategy.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		err := repo.UpdateStatus(ctx, strategy.ID, 1, true)

		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, strategy.ID, 1)
		require.NoError(t, err)
		assert.True(t, found.Active)
		assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("deactivate sets active=false and updates updated_at", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		originalUpdatedAt := strategy.UpdatedAt

		time.Sleep(10 * time.Millisecond)

		err := repo.UpdateStatus(ctx, strategy.ID, 1, false)

		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, strategy.ID, 1)
		require.NoError(t, err)
		assert.False(t, found.Active)
		assert.True(t, found.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("not found returns ErrStrategyNotFound", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		err := repo.UpdateStatus(ctx, 99999, 1, true)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
	})

	t.Run("foreign user ID returns ErrStrategyNotFound", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewPostgresStrategyRepository(db)
		ctx := context.Background()

		strategy := &Strategy{
			UserID:       1,
			Name:         "User 1 Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}
		require.NoError(t, repo.Create(ctx, strategy))

		err := repo.UpdateStatus(ctx, strategy.ID, 2, false)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
	})
}

func stringPtr(s string) *string {
	return &s
}
