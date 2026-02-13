package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/opinedajr/micro-stakes-api/internal/infrastructure/identity"
	customValidator "github.com/opinedajr/micro-stakes-api/internal/shared/validator"
)

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
	RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error)
	Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error)
}

type authService struct {
	repo             UserRepository
	identityProvider identity.IdentityProvider
	logger           *slog.Logger
	validator        *validator.Validate
}

func NewAuthService(repo UserRepository, identityProvider identity.IdentityProvider, logger *slog.Logger) AuthService {
	v := validator.New()
	_ = customValidator.RegisterCustomValidators(v)
	return &authService{
		repo:             repo,
		identityProvider: identityProvider,
		logger:           logger,
		validator:        v,
	}
}

func (s *authService) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	if err := s.validator.Struct(input); err != nil {
		s.logger.Error("validation failed", "error", err)
		return nil, WrapError(ErrValidationFailed, err.Error())
	}

	existingUser, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		s.logger.Error("failed to check existing user", "email", input.Email, "error", err)
		return nil, WrapError(ErrDatabaseError, "failed to check existing user")
	}
	if existingUser != nil {
		s.logger.Warn("user already exists", "email", input.Email)
		return nil, ErrUserAlreadyExists
	}

	identityID, err := s.identityProvider.CreateUser(ctx, input.FirstName, input.LastName, input.Email, input.Password)
	if err != nil {
		s.logger.Error("failed to create user in identity provider", "email", input.Email, "error", err)
		return nil, WrapError(ErrIdentityProviderError, "failed to create user in identity provider")
	}

	fullName := fmt.Sprintf("%s %s", input.FirstName, input.LastName)
	user := &User{
		FullName:        fullName,
		Email:           input.Email,
		IdentityID:      identityID,
		IdentityAdapter: IdentityAdapterKeycloak,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("failed to create user in database", "email", input.Email, "error", err)
		return nil, WrapError(ErrDatabaseError, "failed to create user in database")
	}

	s.logger.Info("user registered successfully", "user_id", user.ID, "email", user.Email)

	return &RegisterOutput{
		ID:       user.ID,
		Email:    user.Email,
		FullName: user.FullName,
		Message:  "User registered successfully",
	}, nil
}

func (s *authService) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	if err := s.validator.Struct(input); err != nil {
		s.logger.Error("validation failed", "error", err)
		return nil, WrapError(ErrValidationFailed, err.Error())
	}

	tokens, err := s.identityProvider.ValidateCredentials(ctx, input.Email, input.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			s.logger.Warn("invalid credentials attempt", "email", input.Email)
			return nil, ErrInvalidCredentials
		}
		if errors.Is(err, ErrTokenGenerationFailed) {
			s.logger.Error("token generation failed", "email", input.Email, "error", err)
			return nil, ErrTokenGenerationFailed
		}
		s.logger.Error("identity provider error during login", "email", input.Email, "error", err)
		return nil, WrapError(ErrIdentityProviderError, "authentication failed")
	}

	s.logger.Info("user logged in successfully", "email", input.Email)

	return &AuthOutput{
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		TokenType:        tokens.TokenType,
		ExpiresIn:        tokens.ExpiresIn,
		RefreshExpiresIn: tokens.RefreshExpiresIn,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, input RefreshTokenInput) (*AuthOutput, error) {
	if err := s.validator.Struct(input); err != nil {
		s.logger.Error("validation failed", "error", err)
		return nil, WrapError(ErrValidationFailed, err.Error())
	}

	tokens, err := s.identityProvider.RefreshToken(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			s.logger.Warn("invalid refresh token attempt")
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("identity provider error during token refresh", "error", err)
		return nil, WrapError(ErrIdentityProviderError, "token refresh failed")
	}

	s.logger.Info("token refreshed successfully")

	return &AuthOutput{
		AccessToken:      tokens.AccessToken,
		RefreshToken:     tokens.RefreshToken,
		TokenType:        tokens.TokenType,
		ExpiresIn:        tokens.ExpiresIn,
		RefreshExpiresIn: tokens.RefreshExpiresIn,
	}, nil
}

func (s *authService) Logout(ctx context.Context, input LogoutInput) (*LogoutOutput, error) {
	if err := s.validator.Struct(input); err != nil {
		s.logger.Error("validation failed", "error", err)
		return nil, WrapError(ErrValidationFailed, err.Error())
	}

	if err := s.identityProvider.RevokeTokens(ctx, input.RefreshToken); err != nil {
		s.logger.Error("identity provider error during logout", "error", err)
		return nil, WrapError(ErrIdentityProviderError, "logout failed")
	}

	s.logger.Info("user logged out successfully")

	return &LogoutOutput{
		Message: "Logged out successfully",
	}, nil
}
