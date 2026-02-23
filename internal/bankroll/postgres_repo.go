package bankroll

import (
	"context"

	"gorm.io/gorm"
)

type postgresBankrollRepository struct {
	db *gorm.DB
}

func NewPostgresBankrollRepository(db *gorm.DB) BankrollRepository {
	return &postgresBankrollRepository{
		db: db,
	}
}

func (r *postgresBankrollRepository) Create(ctx context.Context, bankroll *Bankroll) error {
	var existingBankroll Bankroll
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND name = ?", bankroll.UserID, bankroll.Name).
		First(&existingBankroll).Error

	if err == nil {
		return ErrBankrollNameExists
	}

	if err != gorm.ErrRecordNotFound {
		return WrapError(ErrDatabaseError, err.Error())
	}

	if err := r.db.WithContext(ctx).Create(bankroll).Error; err != nil {
		return WrapError(ErrDatabaseError, err.Error())
	}
	return nil
}

func (r *postgresBankrollRepository) Update(ctx context.Context, bankroll *Bankroll) error {
	var existingBankroll Bankroll
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", bankroll.ID, bankroll.UserID).
		First(&existingBankroll).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrBankrollNotFound
		}
		return WrapError(ErrDatabaseError, err.Error())
	}

	var otherBankroll Bankroll
	err = r.db.WithContext(ctx).
		Where("user_id = ? AND name = ? AND id != ?", bankroll.UserID, bankroll.Name, bankroll.ID).
		First(&otherBankroll).Error

	if err == nil {
		return ErrBankrollNameExists
	}

	if err != gorm.ErrRecordNotFound {
		return WrapError(ErrDatabaseError, err.Error())
	}

	if err := r.db.WithContext(ctx).Model(&existingBankroll).Updates(map[string]interface{}{
		"name":                  bankroll.Name,
		"currency":              bankroll.Currency,
		"start_date":            bankroll.StartDate,
		"commission_percentage": bankroll.CommissionPercentage,
	}).Error; err != nil {
		return WrapError(ErrDatabaseError, err.Error())
	}
	return nil
}

func (r *postgresBankrollRepository) ListByUserID(ctx context.Context, userID uint) ([]*Bankroll, error) {
	var bankrolls []*Bankroll
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&bankrolls).Error
	if err != nil {
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return bankrolls, nil
}

func (r *postgresBankrollRepository) FindByID(ctx context.Context, id uint, userID uint) (*Bankroll, error) {
	var bankroll Bankroll
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&bankroll).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrBankrollNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &bankroll, nil
}

func (r *postgresBankrollRepository) Reset(ctx context.Context, id uint, userID uint) error {
	result := r.db.WithContext(ctx).Model(&Bankroll{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"initial_balance": 0,
			"current_balance": 0,
		})

	if result.Error != nil {
		return WrapError(ErrDatabaseError, result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return ErrBankrollNotFound
	}
	return nil
}
