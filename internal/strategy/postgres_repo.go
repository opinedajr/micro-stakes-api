package strategy

import (
	"context"

	"gorm.io/gorm"
)

type postgresStrategyRepository struct {
	db *gorm.DB
}

func NewPostgresStrategyRepository(db *gorm.DB) StrategyRepository {
	return &postgresStrategyRepository{
		db: db,
	}
}

func (r *postgresStrategyRepository) Create(ctx context.Context, strategy *Strategy) error {
	if err := r.db.WithContext(ctx).Create(strategy).Error; err != nil {
		return WrapError(ErrDatabaseError, err.Error())
	}
	return nil
}

func (r *postgresStrategyRepository) Update(ctx context.Context, strategy *Strategy) error {
	result := r.db.WithContext(ctx).Model(&Strategy{}).
		Where("id = ? AND user_id = ?", strategy.ID, strategy.UserID).
		Updates(map[string]interface{}{
			"name":          strategy.Name,
			"description":   strategy.Description,
			"default_stake": strategy.DefaultStake,
			"type":          strategy.Type,
			"active":        strategy.Active,
		})

	if result.Error != nil {
		return WrapError(ErrDatabaseError, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return ErrStrategyNotFound
	}
	return nil
}

func (r *postgresStrategyRepository) UpdateStatus(ctx context.Context, id uint, userID uint, active bool) error {
	result := r.db.WithContext(ctx).Model(&Strategy{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("active", active)

	if result.Error != nil {
		return WrapError(ErrDatabaseError, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return ErrStrategyNotFound
	}
	return nil
}

func (r *postgresStrategyRepository) FindByID(ctx context.Context, id uint, userID uint) (*Strategy, error) {
	var strategy Strategy
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&strategy).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrStrategyNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &strategy, nil
}

func (r *postgresStrategyRepository) ListByUserID(ctx context.Context, userID uint, page int, pageSize int) ([]*Strategy, int64, error) {
	var strategies []*Strategy
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).Model(&Strategy{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	err = r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&strategies).Error

	if err != nil {
		return nil, 0, WrapError(ErrDatabaseError, err.Error())
	}

	return strategies, total, nil
}
