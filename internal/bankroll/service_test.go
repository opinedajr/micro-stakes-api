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

func TestUpdateBankroll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		existingBankroll := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Old Name",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 5.0,
		}

		updatedBankroll := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Updated Name",
			Currency:             CurrencyUSD,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 3.0,
		}

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(existingBankroll, nil).Once()
		mockRepo.On("Update", ctx, mock.MatchedBy(func(b *Bankroll) bool {
			return b.ID == bankrollID &&
				b.UserID == userID &&
				b.Name == "Updated Name" &&
				b.Currency == CurrencyUSD &&
				b.CommissionPercentage == 3.0 &&
				b.InitialBalance == 1000.00 &&
				b.CurrentBalance == 1000.00
		})).Return(nil).Once()
		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(updatedBankroll, nil).Once()

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyUSD,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Updated Name", output.Name)
		assert.Equal(t, CurrencyUSD, output.Currency)
		assert.Equal(t, 3.0, output.CommissionPercentage)
		assert.Equal(t, 1000.00, output.InitialBalance)
		assert.Equal(t, 1000.00, output.CurrentBalance)
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error - invalid currency", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             "INVALID",
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidCurrency)
		mockRepo.AssertNotCalled(t, "FindByID")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("validation error - invalid commission", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 150.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrInvalidCommission)
		mockRepo.AssertNotCalled(t, "FindByID")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("validation error - invalid date format", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyBRL,
			StartDate:            "invalid-date",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrValidationFailed)
		mockRepo.AssertNotCalled(t, "FindByID")
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("bankroll not found", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrBankrollNotFound).Once()

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(2)
		bankrollID := uint(1)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrBankrollNotFound).Once()

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		mockRepo.AssertExpectations(t)
		mockRepo.AssertNotCalled(t, "Update")
	})

	t.Run("repository error - duplicate name", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		existingBankroll := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Old Name",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 5.0,
		}

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(existingBankroll, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*bankroll.Bankroll")).Return(ErrBankrollNameExists).Once()

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

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
		bankrollID := uint(1)

		existingBankroll := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Old Name",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 5.0,
		}

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(existingBankroll, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*bankroll.Bankroll")).Return(ErrDatabaseError).Once()

		input := UpdateBankrollInput{
			Name:                 "Updated Name",
			Currency:             CurrencyBRL,
			StartDate:            "2026-02-01",
			CommissionPercentage: 3.0,
		}

		output, err := service.UpdateBankroll(ctx, userID, bankrollID, input)

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

func TestListBankrolls(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		bankrolls := []*Bankroll{
			{
				ID:                   1,
				UserID:               userID,
				Name:                 "Bankroll 1",
				Currency:             CurrencyBRL,
				InitialBalance:       1000.00,
				CurrentBalance:       1000.00,
				CommissionPercentage: 5.0,
			},
			{
				ID:                   2,
				UserID:               userID,
				Name:                 "Bankroll 2",
				Currency:             CurrencyUSD,
				InitialBalance:       500.00,
				CurrentBalance:       500.00,
				CommissionPercentage: 3.0,
			},
		}

		mockRepo.On("ListByUserID", ctx, userID).Return(bankrolls, nil).Once()

		outputs, err := service.ListBankrolls(ctx, userID)

		assert.NoError(t, err)
		assert.Len(t, outputs, 2)
		assert.Equal(t, "Bankroll 1", outputs[0].Name)
		assert.Equal(t, "Bankroll 2", outputs[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		mockRepo.On("ListByUserID", ctx, userID).Return([]*Bankroll{}, nil).Once()

		outputs, err := service.ListBankrolls(ctx, userID)

		assert.NoError(t, err)
		assert.Empty(t, outputs)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)

		mockRepo.On("ListByUserID", ctx, userID).Return(nil, ErrDatabaseError).Once()

		outputs, err := service.ListBankrolls(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, outputs)
		assert.ErrorIs(t, err, ErrDatabaseError)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetBankroll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		bankroll := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 5.0,
		}

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(bankroll, nil).Once()

		output, err := service.GetBankroll(ctx, userID, bankrollID)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Main Bankroll", output.Name)
		assert.Equal(t, CurrencyBRL, output.Currency)
		assert.Equal(t, 1000.00, output.InitialBalance)
		assert.Equal(t, 1000.00, output.CurrentBalance)
		mockRepo.AssertExpectations(t)
	})

	t.Run("bankroll not found", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(999)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrBankrollNotFound).Once()

		output, err := service.GetBankroll(ctx, userID, bankrollID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(2)
		bankrollID := uint(1)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrBankrollNotFound).Once()

		output, err := service.GetBankroll(ctx, userID, bankrollID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrDatabaseError).Once()

		output, err := service.GetBankroll(ctx, userID, bankrollID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrDatabaseError)
		mockRepo.AssertExpectations(t)
	})
}

func TestResetBankroll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		beforeReset := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 5.0,
		}

		afterReset := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       0.0,
			CurrentBalance:       0.0,
			CommissionPercentage: 5.0,
		}

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(beforeReset, nil).Once()
		mockRepo.On("Reset", ctx, bankrollID, userID).Return(nil).Once()
		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(afterReset, nil).Once()

		output, err := service.ResetBankroll(ctx, userID, bankrollID)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "Main Bankroll", output.Name)
		assert.Equal(t, CurrencyBRL, output.Currency)
		assert.Equal(t, 0.0, output.InitialBalance)
		assert.Equal(t, 0.0, output.CurrentBalance)
		mockRepo.AssertExpectations(t)
	})

	t.Run("bankroll not found", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(999)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrBankrollNotFound).Once()

		output, err := service.ResetBankroll(ctx, userID, bankrollID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		mockRepo.AssertNotCalled(t, "Reset")
		mockRepo.AssertExpectations(t)
	})

	t.Run("unauthorized - different user", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(2)
		bankrollID := uint(1)

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(nil, ErrBankrollNotFound).Once()

		output, err := service.ResetBankroll(ctx, userID, bankrollID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrBankrollNotFound)
		mockRepo.AssertNotCalled(t, "Reset")
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := new(MockBankrollRepository)
		logger := slog.Default()
		service := NewBankrollService(mockRepo, logger)

		ctx := context.Background()
		userID := uint(1)
		bankrollID := uint(1)

		bankroll := &Bankroll{
			ID:                   bankrollID,
			UserID:               userID,
			Name:                 "Main Bankroll",
			Currency:             CurrencyBRL,
			InitialBalance:       1000.00,
			CurrentBalance:       1000.00,
			CommissionPercentage: 5.0,
		}

		mockRepo.On("FindByID", ctx, bankrollID, userID).Return(bankroll, nil).Once()
		mockRepo.On("Reset", ctx, bankrollID, userID).Return(ErrDatabaseError).Once()

		output, err := service.ResetBankroll(ctx, userID, bankrollID)

		assert.Error(t, err)
		assert.Nil(t, output)
		assert.ErrorIs(t, err, ErrDatabaseError)
		mockRepo.AssertExpectations(t)
	})
}
