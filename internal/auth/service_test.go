package auth

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/opinedajr/micro-stakes-api/internal/infrastructure/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockUserRepository) FindByIdentityID(ctx context.Context, identityID string, adapter IdentityAdapter) (*User, error) {
	args := m.Called(ctx, identityID, adapter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

type MockIdentityProvider struct {
	mock.Mock
}

func (m *MockIdentityProvider) CreateUser(ctx context.Context, firstName, lastName, email, password string) (string, error) {
	args := m.Called(ctx, firstName, lastName, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockIdentityProvider) ValidateCredentials(ctx context.Context, email, password string) (*identity.AuthTokens, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.AuthTokens), args.Error(1)
}

func (m *MockIdentityProvider) RefreshToken(ctx context.Context, refreshToken string) (*identity.AuthTokens, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*identity.AuthTokens), args.Error(1)
}

func (m *MockIdentityProvider) RevokeTokens(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name          string
		input         RegisterInput
		mockRepoSetup func(*MockUserRepository)
		mockIDPSetup  func(*MockIdentityProvider)
		expectError   bool
		errorType     error
	}{
		{
			name: "success - valid registration",
			input: RegisterInput{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Password:  "SecureP@ss123",
			},
			mockRepoSetup: func(repo *MockUserRepository) {
				repo.On("FindByEmail", ctx, "john.doe@example.com").Return(nil, ErrUserNotFound)
				repo.On("CreateUser", ctx, mock.AnythingOfType("*auth.User")).Return(nil)
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("CreateUser", ctx, "John", "Doe", "john.doe@example.com", "SecureP@ss123").Return("keycloak-user-id-123", nil)
			},
			expectError: false,
		},
		{
			name: "error - user already exists",
			input: RegisterInput{
				FirstName: "Jane",
				LastName:  "Smith",
				Email:     "jane.smith@example.com",
				Password:  "AnotherP@ss456",
			},
			mockRepoSetup: func(repo *MockUserRepository) {
				existingUser := &User{
					ID:       1,
					Email:    "jane.smith@example.com",
					FullName: "Jane Smith",
				}
				repo.On("FindByEmail", ctx, "jane.smith@example.com").Return(existingUser, nil)
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
			},
			expectError: true,
			errorType:   ErrUserAlreadyExists,
		},
		{
			name: "error - identity provider failure",
			input: RegisterInput{
				FirstName: "Bob",
				LastName:  "Johnson",
				Email:     "bob.johnson@example.com",
				Password:  "ValidP@ss789",
			},
			mockRepoSetup: func(repo *MockUserRepository) {
				repo.On("FindByEmail", ctx, "bob.johnson@example.com").Return(nil, ErrUserNotFound)
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("CreateUser", ctx, "Bob", "Johnson", "bob.johnson@example.com", "ValidP@ss789").Return("", errors.New("keycloak connection failed"))
			},
			expectError: true,
			errorType:   ErrIdentityProviderError,
		},
		{
			name: "error - database error during user creation",
			input: RegisterInput{
				FirstName: "Alice",
				LastName:  "Brown",
				Email:     "alice.brown@example.com",
				Password:  "TestP@ss321",
			},
			mockRepoSetup: func(repo *MockUserRepository) {
				repo.On("FindByEmail", ctx, "alice.brown@example.com").Return(nil, ErrUserNotFound)
				repo.On("CreateUser", ctx, mock.AnythingOfType("*auth.User")).Return(errors.New("database connection lost"))
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("CreateUser", ctx, "Alice", "Brown", "alice.brown@example.com", "TestP@ss321").Return("keycloak-user-id-456", nil)
			},
			expectError: true,
			errorType:   ErrDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockIDP := new(MockIdentityProvider)

			tt.mockRepoSetup(mockRepo)
			tt.mockIDPSetup(mockIDP)

			service := NewAuthService(mockRepo, mockIDP, logger)

			output, err := service.Register(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, output)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.input.Email, output.Email)
				assert.NotNil(t, output.ID)
				assert.NotEmpty(t, output.FullName)
			}

			mockRepo.AssertExpectations(t)
			mockIDP.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name         string
		input        LoginInput
		mockIDPSetup func(*MockIdentityProvider)
		expectError  bool
		errorType    error
	}{
		{
			name: "success - valid credentials",
			input: LoginInput{
				Email:    "john.doe@example.com",
				Password: "SecureP@ss123",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				tokens := &identity.AuthTokens{
					AccessToken:      "access-token-123",
					RefreshToken:     "refresh-token-456",
					TokenType:        "Bearer",
					ExpiresIn:        900,
					RefreshExpiresIn: 604800,
				}
				idp.On("ValidateCredentials", ctx, "john.doe@example.com", "SecureP@ss123").Return(tokens, nil)
			},
			expectError: false,
		},
		{
			name: "error - invalid credentials",
			input: LoginInput{
				Email:    "john.doe@example.com",
				Password: "WrongPassword",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("ValidateCredentials", ctx, "john.doe@example.com", "WrongPassword").Return(nil, ErrInvalidCredentials)
			},
			expectError: true,
			errorType:   ErrInvalidCredentials,
		},
		{
			name: "error - identity provider failure",
			input: LoginInput{
				Email:    "jane.smith@example.com",
				Password: "AnotherP@ss456",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("ValidateCredentials", ctx, "jane.smith@example.com", "AnotherP@ss456").Return(nil, errors.New("keycloak unavailable"))
			},
			expectError: true,
			errorType:   ErrIdentityProviderError,
		},
		{
			name: "error - token generation failed",
			input: LoginInput{
				Email:    "bob.johnson@example.com",
				Password: "ValidP@ss789",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("ValidateCredentials", ctx, "bob.johnson@example.com", "ValidP@ss789").Return(nil, ErrTokenGenerationFailed)
			},
			expectError: true,
			errorType:   ErrTokenGenerationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIDP := new(MockIdentityProvider)
			tt.mockIDPSetup(mockIDP)

			mockRepo := new(MockUserRepository)
			service := NewAuthService(mockRepo, mockIDP, logger)

			output, err := service.Login(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, output)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotEmpty(t, output.AccessToken)
				assert.NotEmpty(t, output.RefreshToken)
				assert.Equal(t, "Bearer", output.TokenType)
				assert.Equal(t, 900, output.ExpiresIn)
				assert.Equal(t, 604800, output.RefreshExpiresIn)
			}

			mockIDP.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name         string
		input        RefreshTokenInput
		mockIDPSetup func(*MockIdentityProvider)
		expectError  bool
		errorType    error
	}{
		{
			name: "success - valid refresh token",
			input: RefreshTokenInput{
				RefreshToken: "valid-refresh-token",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				tokens := &identity.AuthTokens{
					AccessToken:      "new-access-token-123",
					RefreshToken:     "new-refresh-token-456",
					TokenType:        "Bearer",
					ExpiresIn:        900,
					RefreshExpiresIn: 604800,
				}
				idp.On("RefreshToken", ctx, "valid-refresh-token").Return(tokens, nil)
			},
			expectError: false,
		},
		{
			name: "error - invalid refresh token",
			input: RefreshTokenInput{
				RefreshToken: "invalid-refresh-token",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("RefreshToken", ctx, "invalid-refresh-token").Return(nil, ErrInvalidCredentials)
			},
			expectError: true,
			errorType:   ErrInvalidCredentials,
		},
		{
			name: "error - identity provider failure",
			input: RefreshTokenInput{
				RefreshToken: "valid-but-fails-token",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("RefreshToken", ctx, "valid-but-fails-token").Return(nil, errors.New("keycloak unavailable"))
			},
			expectError: true,
			errorType:   ErrIdentityProviderError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIDP := new(MockIdentityProvider)
			tt.mockIDPSetup(mockIDP)

			mockRepo := new(MockUserRepository)
			service := NewAuthService(mockRepo, mockIDP, logger)

			output, err := service.RefreshToken(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, output)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotEmpty(t, output.AccessToken)
				assert.NotEmpty(t, output.RefreshToken)
				assert.Equal(t, "Bearer", output.TokenType)
				assert.Equal(t, 900, output.ExpiresIn)
				assert.Equal(t, 604800, output.RefreshExpiresIn)
			}

			mockIDP.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tests := []struct {
		name         string
		input        LogoutInput
		mockIDPSetup func(*MockIdentityProvider)
		expectError  bool
		errorType    error
	}{
		{
			name: "success - valid logout",
			input: LogoutInput{
				RefreshToken: "valid-refresh-token",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("RevokeTokens", ctx, "valid-refresh-token").Return(nil)
			},
			expectError: false,
		},
		{
			name: "success - idempotent logout (already revoked token)",
			input: LogoutInput{
				RefreshToken: "already-revoked-token",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("RevokeTokens", ctx, "already-revoked-token").Return(nil)
			},
			expectError: false,
		},
		{
			name: "error - identity provider failure",
			input: LogoutInput{
				RefreshToken: "valid-but-fails-token",
			},
			mockIDPSetup: func(idp *MockIdentityProvider) {
				idp.On("RevokeTokens", ctx, "valid-but-fails-token").Return(errors.New("keycloak unavailable"))
			},
			expectError: true,
			errorType:   ErrIdentityProviderError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockIDP := new(MockIdentityProvider)
			tt.mockIDPSetup(mockIDP)

			mockRepo := new(MockUserRepository)
			service := NewAuthService(mockRepo, mockIDP, logger)

			output, err := service.Logout(ctx, tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, output)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, "Logged out successfully", output.Message)
			}

			mockIDP.AssertExpectations(t)
		})
	}
}
