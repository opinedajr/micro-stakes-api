package bankroll

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
)

type MockBankrollRepository struct {
	mock.Mock
}

func (m *MockBankrollRepository) Create(ctx context.Context, bankroll *Bankroll) error {
	args := m.Called(ctx, bankroll)
	return args.Error(0)
}

func (m *MockBankrollRepository) Update(ctx context.Context, bankroll *Bankroll) error {
	args := m.Called(ctx, bankroll)
	return args.Error(0)
}

func (m *MockBankrollRepository) ListByUserID(ctx context.Context, userID uint) ([]*Bankroll, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Bankroll), args.Error(1)
}

func (m *MockBankrollRepository) FindByID(ctx context.Context, id uint, userID uint) (*Bankroll, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Bankroll), args.Error(1)
}

func (m *MockBankrollRepository) Reset(ctx context.Context, id uint, userID uint) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func TestCreateBankroll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*bankroll.Bankroll")).Return(nil).Once()

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Main Bankroll", output.Name)
		assert.Equal(t, CurrencyBRL, output.Currency)
		assert.Equal(t, 1000.00, output.InitialBalance)
		assert.Equal(t, 1000.00, output.CurrentBalance)
		assert.Equal(t, 5.0, output.CommissionPercentage)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error - invalid currency", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             "INVALID",
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidCurrency)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - negative balance", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       -100.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrNegativeBalance)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - invalid commission", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 150.0,
		}

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidCommission)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("validation error - invalid date format", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "invalid-date",
			CommissionPercentage: 5.0,
		}

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrValidationFailed)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository error - duplicate name", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*bankroll.Bankroll")).Return(ErrBankrollNameExists).Once()

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNameExists)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error - database error", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		input := CreateBankrollInput{
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			StartDate:            "2026-02-01",
			CommissionPercentage: 5.0,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*bankroll.Bankroll")).Return(ErrDatabaseError).Once()

		output, err := service.CreateBankroll(ctx, userID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrDatabaseError)
		mockRepo.AssertExpectations(t)
	})
}

func TestParseDate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dateStr := "2026-02-01"
		parsed, err := parseDate(dateStr)

		assert.NoError(t, err)
		assert.Equal(t, 2026, parsed.Year())
		assert.Equal(t, time.February, parsed.Month())
		assert.Equal(t, 1, parsed.Day())
	})

	t.Run("invalid format", func(t *testing.T) {
		dateStr := "invalid-date"
		parsed, err := parseDate(dateStr)

		assert.Error(t, err)
		assert.Equal(t, time.Time{}, parsed)
	})
}
