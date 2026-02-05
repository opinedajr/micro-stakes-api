package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidators(v *validator.Validate) error {
	if err := v.RegisterValidation("password", validatePassword); err != nil {
		return err
	}
	return nil
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)

	return hasUpper && hasLower && hasDigit
}
