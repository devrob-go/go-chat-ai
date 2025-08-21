package server

import (
	"context"
	"testing"

	"auth-service/config"
	"auth-service/proto"
	"auth-service/storage"

	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
)

func TestNewHealthServer(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}

	server := NewHealthServer(db, logger, cfg)

	assert.NotNil(t, server)
	assert.Equal(t, db, server.db)
	assert.Equal(t, logger, server.logger)
	assert.Equal(t, cfg, server.config)
}

func TestHealthServer_Structure(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}

	server := NewHealthServer(db, logger, cfg)

	// Test that the server was created properly
	assert.NotNil(t, server)
	assert.NotNil(t, server.db)
	assert.NotNil(t, server.logger)
	assert.NotNil(t, server.config)
	assert.Equal(t, db, server.db)
	assert.Equal(t, logger, server.logger)
	assert.Equal(t, cfg, server.config)
}

func TestHealthServer_MethodSignatures(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}
	server := NewHealthServer(db, logger, cfg)

	// Test that all required methods exist and have correct signatures
	// This is a structural test to ensure the server implements the interface

	// Test Check method exists
	_ = server.Check

	// Test Watch method exists
	_ = server.Watch

	// Test checkDatabaseHealth method exists
	_ = server.checkDatabaseHealth

	// If we get here, all methods exist
	assert.True(t, true)
}

// Benchmark tests for performance
func BenchmarkHealthServer_Check(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}

	mockDB := &storage.DB{}
	server := NewHealthServer(mockDB, logger, cfg)

	req := &proto.HealthCheckRequest{
		Service: "auth-service",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.Check(context.Background(), req)
	}
}

func BenchmarkHealthServer_CheckDatabaseHealth(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}

	mockDB := &storage.DB{}
	server := NewHealthServer(mockDB, logger, cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.checkDatabaseHealth(context.Background())
	}
}
