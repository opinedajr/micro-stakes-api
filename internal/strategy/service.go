package strategy

import (
	"context"

	"log/slog"
	"math"
)

type StrategyService interface {
	CreateStrategy(ctx context.Context, userID uint, input CreateStrategyInput) (*StrategyOutput, error)
	ListStrategies(ctx context.Context, userID uint, page int, pageSize int) (*StrategyListOutput, error)
	GetStrategy(ctx context.Context, userID uint, strategyID uint) (*StrategyOutput, error)
	UpdateStrategy(ctx context.Context, userID uint, strategyID uint, input UpdateStrategyInput) (*StrategyOutput, error)
	UpdateStrategyStatus(ctx context.Context, userID uint, strategyID uint, active bool) (*StrategyOutput, error)
}

type strategyService struct {
	repo   StrategyRepository
	logger *slog.Logger
}

func NewStrategyService(repo StrategyRepository, logger *slog.Logger) StrategyService {
	return &strategyService{
		repo:   repo,
		logger: logger,
	}
}

func (s *strategyService) CreateStrategy(ctx context.Context, userID uint, input CreateStrategyInput) (*StrategyOutput, error) {
	if input.DefaultStake < 0.01 {
		s.logger.Error("invalid default stake", "default_stake", input.DefaultStake, "user_id", userID)
		return nil, ErrInvalidDefaultStake
	}

	if input.Type != StrategyTypeBack && input.Type != StrategyTypeLay {
		s.logger.Error("invalid strategy type", "type", input.Type, "user_id", userID)
		return nil, ErrInvalidStrategyType
	}

	active := true
	if input.Active != nil {
		active = *input.Active
	}

	strategy := &Strategy{
		UserID:       userID,
		Name:         input.Name,
		Description:  input.Description,
		DefaultStake: input.DefaultStake,
		Type:         input.Type,
		Active:       active,
	}

	if err := s.repo.Create(ctx, strategy); err != nil {
		s.logger.Error("failed to create strategy", "error", err, "user_id", userID, "name", input.Name)
		return nil, err
	}

	s.logger.Info("strategy created", "user_id", userID, "strategy_id", strategy.ID, "name", input.Name, "type", input.Type, "active", active, "outcome", "success")

	return toStrategyOutput(strategy), nil
}

func (s *strategyService) ListStrategies(ctx context.Context, userID uint, page int, pageSize int) (*StrategyListOutput, error) {
	if page < 1 {
		page = 1
	}

	if pageSize < 1 {
		pageSize = 20
	}

	if pageSize > 100 {
		pageSize = 100
	}

	strategies, total, err := s.repo.ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		s.logger.Error("failed to list strategies", "error", err, "user_id", userID)
		return nil, err
	}

	outputs := make([]*StrategyOutput, len(strategies))
	for i, s := range strategies {
		outputs[i] = toStrategyOutput(s)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &StrategyListOutput{
		Data:       outputs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *strategyService) GetStrategy(ctx context.Context, userID uint, strategyID uint) (*StrategyOutput, error) {
	strategy, err := s.repo.FindByID(ctx, strategyID, userID)
	if err != nil {
		s.logger.Error("strategy not found", "error", err, "user_id", userID, "strategy_id", strategyID)
		return nil, err
	}

	return toStrategyOutput(strategy), nil
}

func (s *strategyService) UpdateStrategy(ctx context.Context, userID uint, strategyID uint, input UpdateStrategyInput) (*StrategyOutput, error) {
	if input.DefaultStake < 0.01 {
		s.logger.Error("invalid default stake", "default_stake", input.DefaultStake, "user_id", userID, "strategy_id", strategyID)
		return nil, ErrInvalidDefaultStake
	}

	if input.Type != StrategyTypeBack && input.Type != StrategyTypeLay {
		s.logger.Error("invalid strategy type", "type", input.Type, "user_id", userID, "strategy_id", strategyID)
		return nil, ErrInvalidStrategyType
	}

	_, err := s.repo.FindByID(ctx, strategyID, userID)
	if err != nil {
		s.logger.Error("strategy not found", "error", err, "user_id", userID, "strategy_id", strategyID)
		return nil, err
	}

	strategy := &Strategy{
		ID:           strategyID,
		UserID:       userID,
		Name:         input.Name,
		Description:  input.Description,
		DefaultStake: input.DefaultStake,
		Type:         input.Type,
		Active:       *input.Active,
	}

	if err := s.repo.Update(ctx, strategy); err != nil {
		s.logger.Error("failed to update strategy", "error", err, "user_id", userID, "strategy_id", strategyID)
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, strategyID, userID)
	if err != nil {
		s.logger.Error("failed to retrieve updated strategy", "error", err, "user_id", userID, "strategy_id", strategyID)
		return nil, err
	}

	s.logger.Info("strategy updated", "user_id", userID, "strategy_id", strategyID, "name", input.Name, "type", input.Type, "active", *input.Active, "outcome", "success")

	return toStrategyOutput(updated), nil
}

func (s *strategyService) UpdateStrategyStatus(ctx context.Context, userID uint, strategyID uint, active bool) (*StrategyOutput, error) {
	if err := s.repo.UpdateStatus(ctx, strategyID, userID, active); err != nil {
		s.logger.Error("failed to update strategy status", "error", err, "user_id", userID, "strategy_id", strategyID, "active", active)
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, strategyID, userID)
	if err != nil {
		s.logger.Error("failed to retrieve updated strategy", "error", err, "user_id", userID, "strategy_id", strategyID)
		return nil, err
	}

	s.logger.Info("strategy status updated", "user_id", userID, "strategy_id", strategyID, "active", active, "outcome", "success")

	return toStrategyOutput(updated), nil
}

func toStrategyOutput(strategy *Strategy) *StrategyOutput {
	return &StrategyOutput{
		ID:           strategy.ID,
		UserID:       strategy.UserID,
		Name:         strategy.Name,
		Description:  strategy.Description,
		DefaultStake: strategy.DefaultStake,
		Type:         strategy.Type,
		Active:       strategy.Active,
		CreatedAt:    strategy.CreatedAt,
		UpdatedAt:    strategy.UpdatedAt,
	}
}
