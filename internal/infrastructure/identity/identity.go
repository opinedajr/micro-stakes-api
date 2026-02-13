package identity

import "context"

type IdentityProvider interface {
	CreateUser(ctx context.Context, firstName, lastName, email, password string) (string, error)
	ValidateCredentials(ctx context.Context, email, password string) (*AuthTokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthTokens, error)
	RevokeTokens(ctx context.Context, refreshToken string) error
}

type AuthTokens struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresIn        int
	RefreshExpiresIn int
}
