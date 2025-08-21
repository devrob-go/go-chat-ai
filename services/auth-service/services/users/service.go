package users

import (
	"auth-service/storage"

	zlog "packages/logger"
)

// UserService handles user operations
type UserService struct {
	DB     *storage.DB
	logger *zlog.Logger
}

// NewUserService creates a new user service
func NewUserService(db *storage.DB, logger *zlog.Logger) *UserService {
	return &UserService{
		DB:     db,
		logger: logger,
	}
}
