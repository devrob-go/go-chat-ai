package server

import (
	"context"
	"fmt"
	"time"

	"auth-service/config"
	"auth-service/proto"
	"auth-service/storage"
	zlog "packages/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HealthServer implements the HealthService gRPC interface
type HealthServer struct {
	proto.UnimplementedHealthServer
	db     *storage.DB
	logger *zlog.Logger
	config *config.Config
}

// NewHealthServer creates a new health server instance
func NewHealthServer(db *storage.DB, logger *zlog.Logger, cfg *config.Config) *HealthServer {
	return &HealthServer{
		db:     db,
		logger: logger,
		config: cfg,
	}
}

// Check provides a health check response
func (s *HealthServer) Check(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	start := time.Now()
	s.logger.Info(ctx, "Health check requested", map[string]any{
		"service": req.Service,
		"time":    start.Format(time.RFC3339),
	})

	// Use configurable timeout for health checks
	healthCtx, cancel := context.WithTimeout(ctx, time.Duration(s.config.HealthCheckTimeout)*time.Second)
	defer cancel()

	// Check database connectivity with a shorter timeout
	dbStart := time.Now()
	dbStatus := s.checkDatabaseHealth(healthCtx)
	dbDuration := time.Since(dbStart)

	// Check if context was cancelled or timed out
	select {
	case <-healthCtx.Done():
		s.logger.Error(ctx, healthCtx.Err(), "Health check timed out or cancelled", 408, map[string]any{
			"timeout_ms": s.config.HealthCheckTimeout * 1000,
			"elapsed_ms": time.Since(start).Milliseconds(),
		})
		return nil, status.Errorf(codes.DeadlineExceeded, "health check timed out after %d seconds", s.config.HealthCheckTimeout)
	case <-ctx.Done():
		// Check if the original context was cancelled (client disconnected)
		s.logger.Warn(ctx, "Client disconnected during health check", map[string]any{
			"error":       ctx.Err().Error(),
			"elapsed_ms":  time.Since(start).Milliseconds(),
		})
		return nil, status.Errorf(codes.Canceled, "client disconnected")
	default:
		// Continue with response
	}

	// Determine overall health
	overallStatus := proto.HealthCheckResponse_SERVING
	if dbStatus != proto.HealthCheckResponse_SERVING {
		overallStatus = proto.HealthCheckResponse_NOT_SERVING
	}

	response := &proto.HealthCheckResponse{
		Status: overallStatus,
	}

	totalDuration := time.Since(start)
	s.logger.Info(ctx, "Health check completed", map[string]any{
		"status":            overallStatus.String(),
		"database_status":   dbStatus.String(),
		"db_duration_ms":    dbDuration.Milliseconds(),
		"total_duration_ms": totalDuration.Milliseconds(),
		"timeout_ms":        s.config.HealthCheckTimeout * 1000,
	})

	return response, nil
}

// Watch provides a streaming health check response
func (s *HealthServer) Watch(req *proto.HealthCheckRequest, stream proto.Health_WatchServer) error {
	ctx := stream.Context()
	s.logger.Info(ctx, "Health watch started", map[string]any{
		"service": req.Service,
	})

	// Send health updates every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info(ctx, "Health watch stopped", map[string]any{
				"reason": "context cancelled",
			})
			return nil
		case <-ticker.C:
			// Check database connectivity
			dbStatus := s.checkDatabaseHealth(ctx)

			// Determine overall health
			overallStatus := proto.HealthCheckResponse_SERVING
			if dbStatus != proto.HealthCheckResponse_SERVING {
				overallStatus = proto.HealthCheckResponse_NOT_SERVING
			}

			response := &proto.HealthCheckResponse{
				Status: overallStatus,
			}

			if err := stream.Send(response); err != nil {
				s.logger.Error(ctx, err, "Failed to send health update", 500)
				return status.Errorf(codes.Internal, "failed to send health update: %v", err)
			}
		}
	}
}

// checkDatabaseHealth checks the database connectivity
func (s *HealthServer) checkDatabaseHealth(ctx context.Context) proto.HealthCheckResponse_ServingStatus {
	if s.db == nil {
		s.logger.Error(ctx, fmt.Errorf("database connection is nil"), "Database connection not initialized", 500, map[string]any{
			"error": "database connection is nil",
		})
		return proto.HealthCheckResponse_NOT_SERVING
	}

	// Use a very short timeout for database ping to prevent hanging
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	pingStart := time.Now()
	if err := s.db.PingContext(pingCtx); err != nil {
		pingDuration := time.Since(pingStart)
		s.logger.Error(ctx, err, "Database health check failed", 500, map[string]any{
			"error":            err.Error(),
			"ping_duration_ms": pingDuration.Milliseconds(),
			"error_type":       fmt.Sprintf("%T", err),
		})
		return proto.HealthCheckResponse_NOT_SERVING
	}

	pingDuration := time.Since(pingStart)
	s.logger.Debug(ctx, "Database ping successful", map[string]any{
		"ping_duration_ms": pingDuration.Milliseconds(),
	})

	return proto.HealthCheckResponse_SERVING
}
