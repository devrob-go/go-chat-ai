package authentication

import (
	"auth-service/storage"

	zlog "packages/logger"
)

// AuthService handles authentication operations
type AuthService struct {
	DB     *storage.DB
	logger *zlog.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(db *storage.DB, logger *zlog.Logger) *AuthService {
	return &AuthService{
		DB:     db,
		logger: logger,
	}
}
