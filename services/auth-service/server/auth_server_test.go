package server

import (
	"testing"
	"time"

	"auth-service/config"
	"auth-service/models"
	"auth-service/services"

	zlog "packages/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthServer(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	service := &services.Service{}

	server := NewAuthServer(service, logger)

	assert.NotNil(t, server)
	assert.Equal(t, service, server.service)
	assert.Equal(t, logger, server.logger)
}

func TestAuthServer_Structure(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Create service with minimal config
	service := &services.Service{
		Config: &config.Config{
			JWTAccessTokenSecret:  "test-access-secret",
			JWTRefreshTokenSecret: "test-refresh-secret",
		},
	}

	server := NewAuthServer(service, logger)

	// Test that the server was created properly
	assert.NotNil(t, server)
	assert.NotNil(t, server.service)
	assert.NotNil(t, server.logger)
	assert.Equal(t, service, server.service)
	assert.Equal(t, logger, server.logger)
}

func TestConvertUserToProto(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	user := &models.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	protoUser := convertUserToProto(user)

	assert.NotNil(t, protoUser)
	assert.Equal(t, userID.String(), protoUser.Id)
	assert.Equal(t, "Test User", protoUser.Name)
	assert.Equal(t, "test@example.com", protoUser.Email)
	assert.NotNil(t, protoUser.CreatedAt)
	assert.NotNil(t, protoUser.UpdatedAt)
}

func TestConvertUserTokenToProto(t *testing.T) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	userToken := &models.UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		AccessExpiresAt:  now.Add(time.Hour),
		RefreshExpiresAt: now.Add(24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        now,
	}

	protoToken := convertUserTokenToProto(userToken)

	assert.NotNil(t, protoToken)
	assert.Equal(t, tokenID.String(), protoToken.Id)
	assert.Equal(t, userID.String(), protoToken.UserId)
	assert.Equal(t, "access-token", protoToken.AccessToken)
	assert.Equal(t, "refresh-token", protoToken.RefreshToken)
	assert.Equal(t, false, protoToken.IsRevoked)
	assert.NotNil(t, protoToken.AccessExpiresAt)
	assert.NotNil(t, protoToken.RefreshExpiresAt)
	assert.NotNil(t, protoToken.CreatedAt)
}

func TestConvertUserToProto_EdgeCases(t *testing.T) {
	// Test with zero values
	user := &models.User{}
	protoUser := convertUserToProto(user)

	assert.NotNil(t, protoUser)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", protoUser.Id) // Zero UUID
	assert.Equal(t, "", protoUser.Name)
	assert.Equal(t, "", protoUser.Email)
	assert.NotNil(t, protoUser.CreatedAt) // timestamppb.New() never returns nil
	assert.NotNil(t, protoUser.UpdatedAt)
}

func TestConvertUserTokenToProto_EdgeCases(t *testing.T) {
	// Test with zero values
	userToken := &models.UserToken{}
	protoToken := convertUserTokenToProto(userToken)

	assert.NotNil(t, protoToken)
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", protoToken.Id)     // Zero UUID
	assert.Equal(t, "00000000-0000-0000-0000-000000000000", protoToken.UserId) // Zero UUID
	assert.Equal(t, "", protoToken.AccessToken)
	assert.Equal(t, "", protoToken.RefreshToken)
	assert.Equal(t, false, protoToken.IsRevoked)
	assert.NotNil(t, protoToken.AccessExpiresAt) // timestamppb.New() never returns nil
	assert.NotNil(t, protoToken.RefreshExpiresAt)
	assert.NotNil(t, protoToken.CreatedAt)
}

func TestAuthServer_MethodSignatures(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	service := &services.Service{}
	server := NewAuthServer(service, logger)

	// Test that all required methods exist and have correct signatures
	// This is a structural test to ensure the server implements the interface

	// Test SignUp method exists
	_ = server.SignUp

	// Test SignIn method exists
	_ = server.SignIn

	// Test SignOut method exists
	_ = server.SignOut

	// Test RefreshToken method exists
	_ = server.RefreshToken

	// Test RevokeToken method exists
	_ = server.RevokeToken

	// Test ListUsers method exists
	_ = server.ListUsers

	// If we get here, all methods exist
	assert.True(t, true)
}

// Benchmark tests for performance
func BenchmarkConvertUserToProto(b *testing.B) {
	userID := uuid.New()
	now := time.Now()
	user := &models.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		CreatedAt: now,
		UpdatedAt: now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertUserToProto(user)
	}
}

func BenchmarkConvertUserTokenToProto(b *testing.B) {
	tokenID := uuid.New()
	userID := uuid.New()
	now := time.Now()
	userToken := &models.UserToken{
		ID:               tokenID,
		UserID:           userID,
		AccessToken:      "access-token",
		RefreshToken:     "refresh-token",
		AccessExpiresAt:  now.Add(time.Hour),
		RefreshExpiresAt: now.Add(24 * time.Hour),
		IsRevoked:        false,
		CreatedAt:        now,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertUserTokenToProto(userToken)
	}
}
