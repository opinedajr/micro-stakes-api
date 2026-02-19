package bankroll

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	customValidator "github.com/opinedajr/micro-stakes-api/internal/shared/validator"
	"log/slog"
)

type BankrollService interface {
	CreateBankroll(ctx context.Context, userID uint, input CreateBankrollInput) (*BankrollOutput, error)
	UpdateBankroll(ctx context.Context, userID uint, bankrollID uint, input UpdateBankrollInput) (*BankrollOutput, error)
	ListBankrolls(ctx context.Context, userID uint) ([]*BankrollOutput, error)
	GetBankroll(ctx context.Context, userID uint, bankrollID uint) (*BankrollOutput, error)
	ResetBankroll(ctx context.Context, userID uint, bankrollID uint) (*BankrollOutput, error)
}

type bankrollService struct {
	repo      BankrollRepository
	logger    *slog.Logger
	validator *validator.Validate
}

func NewBankrollService(repo BankrollRepository, logger *slog.Logger) BankrollService {
	v := validator.New()
	_ = customValidator.RegisterCustomValidators(v)
	return &bankrollService{
		repo:      repo,
		logger:    logger,
		validator: v,
	}
}

func (s *bankrollService) CreateBankroll(ctx context.Context, userID uint, input CreateBankrollInput) (*BankrollOutput, error) {
	if err := s.validator.Struct(input); err != nil {
		s.logger.Error("validation failed", "error", err, "user_id", userID)
		return nil, WrapError(ErrValidationFailed, err.Error())
	}

	if input.InitialBalance < 0 {
		s.logger.Error("negative balance", "initial_balance", input.InitialBalance, "user_id", userID)
		return nil, ErrNegativeBalance
	}

	if input.CommissionPercentage < 0 || input.CommissionPercentage > 100 {
		s.logger.Error("invalid commission", "commission_percentage", input.CommissionPercentage, "user_id", userID)
		return nil, ErrInvalidCommission
	}

	validCurrencies := map[Currency]bool{
		CurrencyBRL: true,
		CurrencyUSD: true,
		CurrencyEUR: true,
		CurrencyBTC: true,
	}

	if !validCurrencies[input.Currency] {
		s.logger.Error("invalid currency", "currency", input.Currency, "user_id", userID)
		return nil, ErrInvalidCurrency
	}

	startDate, err := parseDate(input.StartDate)
	if err != nil {
		s.logger.Error("failed to parse start date", "error", err, "user_id", userID, "start_date", input.StartDate)
		return nil, WrapError(ErrValidationFailed, "invalid date format")
	}

	bankroll := &Bankroll{
		UserID:               userID,
		Name:                 input.Name,
		Currency:             input.Currency,
		InitialBalance:       input.InitialBalance,
		CurrentBalance:       input.InitialBalance,
		StartDate:            startDate,
		CommissionPercentage: input.CommissionPercentage,
	}

	if err := s.repo.Create(ctx, bankroll); err != nil {
		s.logger.Error("failed to create bankroll", "error", err, "user_id", userID, "name", input.Name)
		return nil, err
	}

	s.logger.Info("bankroll created", "user_id", userID, "bankroll_id", bankroll.ID, "name", input.Name, "currency", input.Currency, "initial_balance", input.InitialBalance)

	return toBankrollOutput(bankroll), nil
}

func (s *bankrollService) UpdateBankroll(ctx context.Context, userID uint, bankrollID uint, input UpdateBankrollInput) (*BankrollOutput, error) {
	panic("not implemented")
}

func (s *bankrollService) ListBankrolls(ctx context.Context, userID uint) ([]*BankrollOutput, error) {
	panic("not implemented")
}

func (s *bankrollService) GetBankroll(ctx context.Context, userID uint, bankrollID uint) (*BankrollOutput, error) {
	panic("not implemented")
}

func (s *bankrollService) ResetBankroll(ctx context.Context, userID uint, bankrollID uint) (*BankrollOutput, error) {
	panic("not implemented")
}

func toBankrollOutput(bankroll *Bankroll) *BankrollOutput {
	return &BankrollOutput{
		ID:                   bankroll.ID,
		Name:                 bankroll.Name,
		Currency:             bankroll.Currency,
		InitialBalance:       bankroll.InitialBalance,
		CurrentBalance:       bankroll.CurrentBalance,
		StartDate:            bankroll.StartDate.Format("2006-01-02"),
		CommissionPercentage: bankroll.CommissionPercentage,
		CreatedAt:            bankroll.CreatedAt,
		UpdatedAt:            bankroll.UpdatedAt,
	}
}

func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
