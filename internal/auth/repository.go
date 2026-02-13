package auth

import (
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	FindByIdentityID(ctx context.Context, identityID string, adapter IdentityAdapter) (*User, error)
}
