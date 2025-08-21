package server

import (
	"auth-service/config"
	"auth-service/proto"
	"auth-service/services"
	"auth-service/storage"
	zlog "packages/logger"

	"google.golang.org/grpc"
)

// RegisterServices registers all gRPC services and their REST gateways
func RegisterServices(grpcServer *grpc.Server, svc *services.Service, logger *zlog.Logger, db *storage.DB, cfg *config.Config) {
	// Register gRPC services
	proto.RegisterAuthServiceServer(grpcServer, NewAuthServer(svc, logger))
	proto.RegisterHealthServer(grpcServer, NewHealthServer(db, logger, cfg))

	// Note: REST gateway handlers are registered in the server.go file
	// when creating the REST gateway to avoid conflicts
}
