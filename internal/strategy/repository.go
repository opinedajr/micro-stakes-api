package strategy

import (
	"context"
)

type StrategyRepository interface {
	Create(ctx context.Context, strategy *Strategy) error
	Update(ctx context.Context, strategy *Strategy) error
	UpdateStatus(ctx context.Context, id uint, userID uint, active bool) error
	FindByID(ctx context.Context, id uint, userID uint) (*Strategy, error)
	ListByUserID(ctx context.Context, userID uint, page int, pageSize int) ([]*Strategy, int64, error)
}
