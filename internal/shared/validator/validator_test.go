package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type TestPassword struct {
	Password string `validate:"password"`
}

func TestValidatePassword(t *testing.T) {
	v := validator.New()
	err := RegisterCustomValidators(v)
	assert.NoError(t, err)

	tests := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{
			name:      "success - valid password with all requirements",
			password:  "ValidP@ss123",
			expectErr: false,
		},
		{
			name:      "success - valid password with special characters",
			password:  "Complex!Pass2024",
			expectErr: false,
		},
		{
			name:      "success - valid password with numbers",
			password:  "Secure456Abc",
			expectErr: false,
		},
		{
			name:      "error - password too short (7 chars)",
			password:  "Short1!",
			expectErr: true,
		},
		{
			name:      "error - password too short (4 chars)",
			password:  "Ab1!",
			expectErr: true,
		},
		{
			name:      "error - no uppercase letter",
			password:  "lowercase123!",
			expectErr: true,
		},
		{
			name:      "error - no lowercase letter",
			password:  "UPPERCASE123!",
			expectErr: true,
		},
		{
			name:      "error - no digit",
			password:  "NoDigitsHere!",
			expectErr: true,
		},
		{
			name:      "error - only lowercase",
			password:  "onlylowercase",
			expectErr: true,
		},
		{
			name:      "error - only uppercase",
			password:  "ONLYUPPERCASE",
			expectErr: true,
		},
		{
			name:      "error - only digits",
			password:  "12345678",
			expectErr: true,
		},
		{
			name:      "error - mixed case but no digit",
			password:  "MixedCaseNoDigit!",
			expectErr: true,
		},
		{
			name:      "error - empty string",
			password:  "",
			expectErr: true,
		},
		{
			name:      "error - exactly 8 chars but missing uppercase",
			password:  "lower12!",
			expectErr: true,
		},
		{
			name:      "success - exactly 8 chars with all requirements",
			password:  "Valid8!C",
			expectErr: false,
		},
		{
			name:      "success - long password",
			password:  "VeryLongP@ssword123456",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testStruct := TestPassword{Password: tt.password}
			err := v.Struct(testStruct)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegisterCustomValidators(t *testing.T) {
	v := validator.New()

	err := RegisterCustomValidators(v)
	assert.NoError(t, err)

	type PasswordTest struct {
		Password string `validate:"password"`
	}

	tests := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{
			name:      "password validation registered - valid",
			password:  "ValidP@ss123",
			expectErr: false,
		},
		{
			name:      "password validation registered - invalid",
			password:  "invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt := PasswordTest{Password: tt.password}
			err := v.Struct(pt)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
