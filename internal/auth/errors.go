package auth

import (
	"errors"
	"fmt"
)

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrValidationFailed      = errors.New("validation failed")
	ErrIdentityProviderError = errors.New("identity provider error")
	ErrDatabaseError         = errors.New("database error")
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrTokenGenerationFailed = errors.New("token generation failed")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
