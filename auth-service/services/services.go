package services

import (
	"auth-service/config"
	auth "auth-service/services/auth"
	"auth-service/services/users"
	"auth-service/storage"

	zlog "packages/logger"
)

// Service encapsulates all business logic services
type Service struct {
	Config *config.Config
	DB     *storage.DB
	User   *users.UserService
	Auth   *auth.AuthService
}

// NewService creates a new service instance
func NewService(db *storage.DB, logger *zlog.Logger, cfg *config.Config) *Service {
	return &Service{
		Config: cfg,
		DB:     db,
		User:   users.NewUserService(db, logger),
		Auth:   auth.NewAuthService(db, logger),
	}
}
