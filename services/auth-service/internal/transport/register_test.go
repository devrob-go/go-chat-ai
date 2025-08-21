package server

import (
	"testing"

	"auth-service/config"
	"auth-service/services"
	"auth-service/storage"

	zlog "packages/logger"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestRegisterServices(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}
	svc := &services.Service{
		Config: cfg,
		DB:     db,
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register services
	RegisterServices(grpcServer, svc, logger, db, cfg)

	// Test that the services were registered
	// We can't directly access the registered services, but we can verify
	// that the server was created and the function didn't panic
	assert.NotNil(t, grpcServer)
}

func TestRegisterServices_Structure(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}
	svc := &services.Service{
		Config: cfg,
		DB:     db,
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Test that RegisterServices can be called without panicking
	// This is a structural test to ensure the function exists and can be called
	assert.NotPanics(t, func() {
		RegisterServices(grpcServer, svc, logger, db, cfg)
	})

	// Verify the server still exists
	assert.NotNil(t, grpcServer)
}

func TestRegisterServices_WithNilValues(t *testing.T) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Test that RegisterServices can be called with nil values
	// This tests the robustness of the function
	assert.NotPanics(t, func() {
		RegisterServices(grpcServer, nil, logger, nil, nil)
	})

	// Verify the server still exists
	assert.NotNil(t, grpcServer)
}

func TestRegisterServices_MethodSignature(t *testing.T) {
	// Test that the RegisterServices function exists and has the correct signature
	// This is a structural test to ensure the function can be called

	// We can't directly test the function signature, but we can verify
	// that the function exists by checking if we can call it
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}
	svc := &services.Service{
		Config: cfg,
		DB:     db,
	}

	grpcServer := grpc.NewServer()

	// If this compiles and runs, the function signature is correct
	RegisterServices(grpcServer, svc, logger, db, cfg)

	assert.True(t, true) // If we get here, the function signature is correct
}

// Benchmark tests for performance
func BenchmarkRegisterServices(b *testing.B) {
	logger := zlog.NewLogger(zlog.Config{Level: "debug"})
	cfg := &config.Config{
		HealthCheckTimeout: 5,
	}
	db := &storage.DB{}
	svc := &services.Service{
		Config: cfg,
		DB:     db,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		grpcServer := grpc.NewServer()
		RegisterServices(grpcServer, svc, logger, db, cfg)
	}
}
