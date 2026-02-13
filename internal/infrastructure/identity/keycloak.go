package identity

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/cenkalti/backoff/v4"
	"github.com/opinedajr/micro-stakes-api/internal/shared/config"
)

type KeycloakAdapter struct {
	client       *gocloak.GoCloak
	config       config.KeycloakConfig
	logger       *slog.Logger
	adminToken   *gocloak.JWT
	tokenExpires time.Time
}

func NewKeycloakAdapter(cfg config.KeycloakConfig, logger *slog.Logger) (IdentityProvider, error) {
	client := gocloak.NewClient(cfg.URL)
	adapter := &KeycloakAdapter{
		client: client,
		config: cfg,
		logger: logger,
	}

	if err := adapter.refreshAdminToken(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to obtain admin token: %w", err)
	}

	return adapter, nil
}

func (k *KeycloakAdapter) refreshAdminToken(ctx context.Context) error {
	token, err := k.client.LoginAdmin(ctx, k.config.AdminUser, k.config.AdminPassword, k.config.AdminRealm)
	if err != nil {
		return err
	}
	k.adminToken = token
	k.tokenExpires = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	return nil
}

func (k *KeycloakAdapter) ensureAdminToken(ctx context.Context) error {
	if time.Now().After(k.tokenExpires.Add(-30 * time.Second)) {
		return k.refreshAdminToken(ctx)
	}
	return nil
}

func (k *KeycloakAdapter) CreateUser(ctx context.Context, firstName, lastName, email, password string) (string, error) {
	var userID string

	operation := func() error {
		if err := k.ensureAdminToken(ctx); err != nil {
			return backoff.Permanent(err)
		}

		enabled := true
		user := gocloak.User{
			Username:  &email,
			Email:     &email,
			FirstName: &firstName,
			LastName:  &lastName,
			Enabled:   &enabled,
		}

		id, err := k.client.CreateUser(ctx, k.adminToken.AccessToken, k.config.Realm, user)
		if err != nil {
			k.logger.Error("failed to create user in Keycloak",
				"email", email,
				"error", err)
			return err
		}
		userID = id

		err = k.client.SetPassword(ctx, k.adminToken.AccessToken, userID, k.config.Realm, password, false)
		if err != nil {
			k.logger.Error("failed to set password in Keycloak",
				"userID", userID,
				"error", err)
			return err
		}

		return nil
	}

	if err := k.retryWithBackoff(operation); err != nil {
		return "", err
	}

	return userID, nil
}

func (k *KeycloakAdapter) ValidateCredentials(ctx context.Context, email, password string) (*AuthTokens, error) {
	var tokens *AuthTokens

	operation := func() error {
		token, err := k.client.Login(ctx, k.config.ClientID, k.config.ClientSecret, k.config.Realm, email, password)
		if err != nil {
			k.logger.Error("failed to validate credentials",
				"email", email,
				"error", err)
			return err
		}

		tokens = &AuthTokens{
			AccessToken:      token.AccessToken,
			RefreshToken:     token.RefreshToken,
			TokenType:        token.TokenType,
			ExpiresIn:        token.ExpiresIn,
			RefreshExpiresIn: token.RefreshExpiresIn,
		}
		return nil
	}

	if err := k.retryWithBackoff(operation); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (k *KeycloakAdapter) RefreshToken(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	var tokens *AuthTokens

	operation := func() error {
		token, err := k.client.RefreshToken(ctx, refreshToken, k.config.ClientID, k.config.ClientSecret, k.config.Realm)
		if err != nil {
			k.logger.Error("failed to refresh token",
				"error", err)
			return err
		}

		tokens = &AuthTokens{
			AccessToken:      token.AccessToken,
			RefreshToken:     token.RefreshToken,
			TokenType:        token.TokenType,
			ExpiresIn:        token.ExpiresIn,
			RefreshExpiresIn: token.RefreshExpiresIn,
		}
		return nil
	}

	if err := k.retryWithBackoff(operation); err != nil {
		return nil, err
	}

	return tokens, nil
}

func (k *KeycloakAdapter) RevokeTokens(ctx context.Context, refreshToken string) error {
	operation := func() error {
		err := k.client.Logout(ctx, k.config.ClientID, k.config.ClientSecret, k.config.Realm, refreshToken)
		if err != nil {
			k.logger.Error("failed to revoke tokens",
				"error", err)
			return err
		}
		return nil
	}

	return k.retryWithBackoff(operation)
}

func (k *KeycloakAdapter) retryWithBackoff(operation func() error) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 5 * time.Second

	return backoff.RetryNotify(
		operation,
		backoff.WithMaxRetries(expBackoff, 3),
		func(err error, duration time.Duration) {
			k.logger.Warn("Keycloak request failed, retrying...",
				"error", err,
				"retry_after", duration)
		},
	)
}
