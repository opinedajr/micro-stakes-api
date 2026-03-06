package strategy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

type MockStrategyRepository struct {
	mock.Mock
}

func (m *MockStrategyRepository) Create(ctx context.Context, strategy *Strategy) error {
	args := m.Called(ctx, strategy)
	return args.Error(0)
}

func (m *MockStrategyRepository) Update(ctx context.Context, strategy *Strategy) error {
	args := m.Called(ctx, strategy)
	return args.Error(0)
}

func (m *MockStrategyRepository) UpdateStatus(ctx context.Context, id uint, userID uint, active bool) error {
	args := m.Called(ctx, id, userID, active)
	return args.Error(0)
}

func (m *MockStrategyRepository) FindByID(ctx context.Context, id uint, userID uint) (*Strategy, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Strategy), args.Error(1)
}

func (m *MockStrategyRepository) ListByUserID(ctx context.Context, userID uint, page int, pageSize int) ([]*Strategy, int64, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]*Strategy), args.Get(1).(int64), args.Error(2)
}

func TestCreateStrategy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		desc := "Back high-value selections at odds > 2.0"
		input := CreateStrategyInput{
			Name:         "Value Back Strategy",
			Description:  &desc,
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*strategy.Strategy")).Return(nil).Once()

		output, err := service.CreateStrategy(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Value Back Strategy", output.Name)
		assert.Equal(t, "Back high-value selections at odds > 2.0", *output.Description)
		assert.Equal(t, 10.00, output.DefaultStake)
		assert.Equal(t, StrategyTypeBack, output.Type)
		assert.True(t, output.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success with nil description", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateStrategyInput{
			Name:         "Lay the Draw",
			Description:  nil,
			DefaultStake: 5.50,
			Type:         StrategyTypeLay,
			Active:       boolPtr(false),
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*strategy.Strategy")).Return(nil).Once()

		output, err := service.CreateStrategy(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Lay the Draw", output.Name)
		assert.Nil(t, output.Description)
		assert.Equal(t, 5.50, output.DefaultStake)
		assert.Equal(t, StrategyTypeLay, output.Type)
		assert.False(t, output.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error - invalid default stake", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 0.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		output, err := service.CreateStrategy(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidDefaultStake)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - invalid strategy type", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         "InvalidType",
			Active:       boolPtr(true),
		}

		output, err := service.CreateStrategy(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidStrategyType)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateStrategyInput{
			Name:         "Test Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*strategy.Strategy")).Return(ErrDatabaseError).Once()

		output, err := service.CreateStrategy(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrDatabaseError)
		mockRepo.AssertExpectations(t)
	})
}

func TestListStrategies(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		strategies := []*Strategy{
			{
				ID:           2,
				UserID:       1,
				Name:         "Strategy 2",
				Description:  stringPtr("Second"),
				DefaultStake: 20.00,
				Type:         StrategyTypeLay,
				Active:       false,
			},
			{
				ID:           1,
				UserID:       1,
				Name:         "Strategy 1",
				Description:  stringPtr("First"),
				DefaultStake: 10.00,
				Type:         StrategyTypeBack,
				Active:       true,
			},
		}

		mockRepo.On("ListByUserID", ctx, userID, 1, 20).Return(strategies, int64(25), nil).Once()

		output, err := service.ListStrategies(ctx, userID, 1, 20)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(25), output.Total)
		assert.Equal(t, 1, output.Page)
		assert.Equal(t, 20, output.PageSize)
		assert.Equal(t, 2, output.TotalPages)
		assert.Len(t, output.Data, 2)
		assert.Equal(t, "Strategy 2", output.Data[0].Name)
		assert.Equal(t, "Strategy 1", output.Data[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		mockRepo.On("ListByUserID", ctx, userID, 1, 20).Return([]*Strategy{}, int64(0), nil).Once()

		output, err := service.ListStrategies(ctx, userID, 1, 20)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(0), output.Total)
		assert.Equal(t, 1, output.Page)
		assert.Equal(t, 20, output.PageSize)
		assert.Equal(t, 0, output.TotalPages)
		assert.Len(t, output.Data, 0)
		mockRepo.AssertExpectations(t)
	})

	t.Run("default page and page_size applied when not provided", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		mockRepo.On("ListByUserID", ctx, userID, 1, 20).Return([]*Strategy{}, int64(0), nil).Once()

		output, err := service.ListStrategies(ctx, userID, 0, 0)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, 1, output.Page)
		assert.Equal(t, 20, output.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("page_size capped at 100", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		mockRepo.On("ListByUserID", ctx, userID, 1, 100).Return([]*Strategy{}, int64(0), nil).Once()

		output, err := service.ListStrategies(ctx, userID, 1, 200)

		assert.NoError(t, err)
		assert.Equal(t, 100, output.PageSize)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetStrategy(t *testing.T) {
	t.Run("success maps to StrategyOutput", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(123)

		strategy := &Strategy{
			ID:           strategyID,
			UserID:       userID,
			Name:         "Value Back Strategy",
			Description:  stringPtr("Back high-value selections at odds > 2.0"),
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}

		mockRepo.On("FindByID", ctx, strategyID, userID).Return(strategy, nil).Once()

		output, err := service.GetStrategy(ctx, userID, strategyID)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, uint(123), output.ID)
		assert.Equal(t, uint(1), output.UserID)
		assert.Equal(t, "Value Back Strategy", output.Name)
		assert.Equal(t, "Back high-value selections at odds > 2.0", *output.Description)
		assert.Equal(t, 10.00, output.DefaultStake)
		assert.Equal(t, StrategyTypeBack, output.Type)
		assert.True(t, output.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo returns ErrStrategyNotFound propagates", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(99999)

		mockRepo.On("FindByID", ctx, strategyID, userID).Return(nil, ErrStrategyNotFound).Once()

		output, err := service.GetStrategy(ctx, userID, strategyID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func TestUpdateStrategy(t *testing.T) {
	t.Run("success maps input to domain model calls repo returns updated StrategyOutput", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(123)
		updatedDesc := "Refined criteria for value bets"

		existingStrategy := &Strategy{
			ID:           strategyID,
			UserID:       userID,
			Name:         "Original Name",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}

		updatedStrategy := &Strategy{
			ID:           strategyID,
			UserID:       userID,
			Name:         "Updated Value Strategy",
			Description:  &updatedDesc,
			DefaultStake: 15.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}

		input := UpdateStrategyInput{
			Name:         "Updated Value Strategy",
			Description:  &updatedDesc,
			DefaultStake: 15.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		mockRepo.On("FindByID", ctx, strategyID, userID).Return(existingStrategy, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*strategy.Strategy")).Return(nil).Once()
		mockRepo.On("FindByID", ctx, strategyID, userID).Return(updatedStrategy, nil).Once()

		output, err := service.UpdateStrategy(ctx, userID, strategyID, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Updated Value Strategy", output.Name)
		assert.Equal(t, "Refined criteria for value bets", *output.Description)
		assert.Equal(t, 15.00, output.DefaultStake)
		assert.Equal(t, StrategyTypeBack, output.Type)
		assert.True(t, output.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo returns ErrStrategyNotFound propagated", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(99999)

		input := UpdateStrategyInput{
			Name:         "Updated Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		mockRepo.On("FindByID", ctx, strategyID, userID).Return(nil, ErrStrategyNotFound).Once()

		output, err := service.UpdateStrategy(ctx, userID, strategyID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid stake returns ErrInvalidDefaultStake", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(123)

		input := UpdateStrategyInput{
			Name:         "Updated Strategy",
			DefaultStake: 0.00,
			Type:         StrategyTypeBack,
			Active:       boolPtr(true),
		}

		output, err := service.UpdateStrategy(ctx, userID, strategyID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidDefaultStake)
		mockRepo.AssertNotCalled(t, "FindByID")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("invalid type returns ErrInvalidStrategyType", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(123)

		input := UpdateStrategyInput{
			Name:         "Updated Strategy",
			DefaultStake: 10.00,
			Type:         "InvalidType",
			Active:       boolPtr(true),
		}

		output, err := service.UpdateStrategy(ctx, userID, strategyID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidStrategyType)
		mockRepo.AssertNotCalled(t, "FindByID")
		mockRepo.AssertNotCalled(t, "Update")
	})
}

func TestUpdateStrategyStatus(t *testing.T) {
	t.Run("success activate calls repo.UpdateStatus with correct args returns updated StrategyOutput", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(123)

		updatedStrategy := &Strategy{
			ID:           strategyID,
			UserID:       userID,
			Name:         "Value Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       true,
		}

		mockRepo.On("UpdateStatus", ctx, strategyID, userID, true).Return(nil).Once()
		mockRepo.On("FindByID", ctx, strategyID, userID).Return(updatedStrategy, nil).Once()

		output, err := service.UpdateStrategyStatus(ctx, userID, strategyID, true)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.True(t, output.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("success deactivate calls repo.UpdateStatus with correct args returns updated StrategyOutput", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(123)

		updatedStrategy := &Strategy{
			ID:           strategyID,
			UserID:       userID,
			Name:         "Value Strategy",
			DefaultStake: 10.00,
			Type:         StrategyTypeBack,
			Active:       false,
		}

		mockRepo.On("UpdateStatus", ctx, strategyID, userID, false).Return(nil).Once()
		mockRepo.On("FindByID", ctx, strategyID, userID).Return(updatedStrategy, nil).Once()

		output, err := service.UpdateStrategyStatus(ctx, userID, strategyID, false)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.False(t, output.Active)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repo returns ErrStrategyNotFound propagated", func(t *testing.T) {
		mockRepo := new(MockStrategyRepository)
		logger := slog.Default()
		service := NewStrategyService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		strategyID := uint(99999)

		mockRepo.On("UpdateStatus", ctx, strategyID, userID, true).Return(ErrStrategyNotFound).Once()

		output, err := service.UpdateStrategyStatus(ctx, userID, strategyID, true)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrStrategyNotFound)
		mockRepo.AssertExpectations(t)
	})
}

func boolPtr(b bool) *bool {
	return &b
}
