package strategy

import (
	"errors"
	"fmt"
)

var (
	ErrStrategyNotFound    = errors.New("strategy not found")
	ErrValidationFailed    = errors.New("validation failed")
	ErrDatabaseError       = errors.New("database error")
	ErrInvalidStrategyType = errors.New("invalid strategy type")
	ErrInvalidDefaultStake = errors.New("invalid default stake")
)

func WrapError(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
