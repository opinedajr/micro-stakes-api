package bankroll

import (
	"errors"
	"fmt"
)

var (
	ErrBankrollNotFound    = errors.New("bankroll not found")
	ErrBankrollNameExists  = errors.New("bankroll name already exists for user")
	ErrValidationFailed    = errors.New("validation failed")
	ErrDatabaseError       = errors.New("database error")
	ErrUnauthorized        = errors.New("unauthorized access to bankroll")
	ErrInvalidCurrency     = errors.New("invalid currency")
	ErrNegativeBalance     = errors.New("balance cannot be negative")
	ErrInvalidCommission   = errors.New("commission percentage must be between 0 and 100")
	ErrCannotModifyBalance = errors.New("cannot modify initial or current balance on update")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
