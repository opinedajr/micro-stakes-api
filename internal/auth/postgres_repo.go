package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type postgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return WrapError(ErrDatabaseError, err.Error())
	}
	return nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &user, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &user, nil
}

func (r *postgresUserRepository) FindByIdentityID(ctx context.Context, identityID string, adapter IdentityAdapter) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("identity_id = ? AND identity_adapter = ?", identityID, adapter).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, WrapError(ErrDatabaseError, err.Error())
	}
	return &user, nil
}
