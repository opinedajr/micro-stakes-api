package bankroll

import (
	"context"
)

type BankrollRepository interface {
	Create(ctx context.Context, bankroll *Bankroll) error
	Update(ctx context.Context, bankroll *Bankroll) error
	ListByUserID(ctx context.Context, userID uint) ([]*Bankroll, error)
	FindByID(ctx context.Context, id uint, userID uint) (*Bankroll, error)
	Reset(ctx context.Context, id uint, userID uint) error
}
