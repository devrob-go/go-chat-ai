package domain

import (
	"context"
	"time"
)

// Credentials represents user login credentials
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserToken represents a user's authentication tokens
type UserToken struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	IsRevoked        bool      `json:"is_revoked"`
	CreatedAt        time.Time `json:"created_at"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User   *User      `json:"user"`
	Tokens *UserToken `json:"tokens"`
}

// TokenResponse represents token-only response
type TokenResponse struct {
	Tokens *UserToken `json:"tokens"`
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	SignUp(name, email, password string) (*AuthResponse, error)
	SignIn(email, password string) (*AuthResponse, error)
	SignOut(accessToken string) error
	RefreshToken(refreshToken string) (*TokenResponse, error)
	RevokeToken(accessToken string) error
	ValidateToken(ctx context.Context, accessToken string, secret string) (*User, error)
}

// TokenRepository defines the interface for token data operations
type TokenRepository interface {
	Create(token *UserToken) error
	GetByAccessToken(accessToken string) (*UserToken, error)
	GetByRefreshToken(refreshToken string) (*UserToken, error)
	Revoke(accessToken string) error
	CleanupExpired() error
}
