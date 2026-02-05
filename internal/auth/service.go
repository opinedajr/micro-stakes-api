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
