package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"auth-service/config"
	"auth-service/proto"
	"auth-service/services"
	"auth-service/storage"
	zlog "packages/logger"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 5 * time.Second
)

// Server holds both gRPC and REST servers and their dependencies
type Server struct {
	logger     *zlog.Logger
	config     *config.Config
	db         *storage.DB
	service    *services.Service
	grpcServer *grpc.Server
	restServer *http.Server
	grpcLis    net.Listener
	restLis    net.Listener
}

// NewServer initializes both gRPC and REST servers with their dependencies
func NewServer(ctx context.Context) (*Server, error) {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger := zlog.NewLogger(zlog.Config{
		Level:      cfg.LogLevel,
		Output:     os.Stdout,
		JSONFormat: cfg.LogJSONFormat,
		AddCaller:  true,
		TimeFormat: time.RFC3339,
	})

	// Create a context with correlation ID for initialization
	ctx = zlog.WithCorrelationID(ctx, "")

	// Initialize database
	logger.Info(ctx, "Initializing database")
	db, err := storage.InitDB(ctx, cfg, logger)
	if err != nil {
		logger.Error(ctx, err, "Failed to initialize database", 500)
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize service
	logger.Info(ctx, "Creating service")
	svc := services.NewService(db, logger, cfg)

	// Create gRPC server with middleware
	metricsMiddleware := NewMetricsMiddleware(logger)
	recoveryMiddleware := NewRecoveryMiddleware(logger)
	rateLimitMiddleware := NewRateLimitMiddleware(logger, cfg)
	securityMiddleware := NewSecurityMiddleware(logger, cfg)

	grpcServer := grpc.NewServer(
		// Recovery middleware should be first to catch panics
		grpc.ChainUnaryInterceptor(
			recoveryMiddleware.UnaryRecoveryInterceptor(),
			securityMiddleware.UnarySecurityInterceptor(),
			metricsMiddleware.UnaryMetricsInterceptor(),
			rateLimitMiddleware.UnaryRateLimitInterceptor(),
			UnaryLoggingInterceptor(logger),
		),
		grpc.ChainStreamInterceptor(
			recoveryMiddleware.StreamRecoveryInterceptor(),
			securityMiddleware.StreamSecurityInterceptor(),
			metricsMiddleware.StreamMetricsInterceptor(),
			rateLimitMiddleware.StreamRateLimitInterceptor(),
			StreamLoggingInterceptor(logger),
		),
	)

	// Register services
	RegisterServices(grpcServer, svc, logger, db, cfg)

	// Enable reflection for development
	if cfg.Environment != config.PRODUCTION_ENV {
		reflection.Register(grpcServer)
	}

	// Create gRPC listener with TLS if enabled
	var grpcLis net.Listener
	if cfg.TLSEnabled {
		// Load TLS configuration
		tlsConfig, err := createTLSConfig(cfg)
		if err != nil {
			logger.Error(ctx, err, "Failed to create TLS configuration", 500)
			return nil, fmt.Errorf("failed to create TLS configuration: %w", err)
		}
		
		// Create TLS listener
		grpcLis, err = tls.Listen("tcp", ":"+cfg.AuthServicePort, tlsConfig)
		if err != nil {
			logger.Error(ctx, err, "Failed to create TLS gRPC listener", 500)
			return nil, fmt.Errorf("failed to create TLS gRPC listener: %w", err)
		}
		logger.Info(ctx, "gRPC server configured with TLS", map[string]any{
			"port": cfg.AuthServicePort,
			"min_tls": cfg.MinTLSVersion,
			"max_tls": cfg.MaxTLSVersion,
		})
	} else {
		// Create plain TCP listener
		grpcLis, err = net.Listen("tcp", ":"+cfg.AuthServicePort)
		if err != nil {
			logger.Error(ctx, err, "Failed to create gRPC listener", 500)
			return nil, fmt.Errorf("failed to create gRPC listener: %w", err)
		}
		logger.Info(ctx, "gRPC server configured without TLS", map[string]any{
			"port": cfg.AuthServicePort,
		})
	}

	// Create REST gateway
	restServer, restLis, err := createRESTGateway(ctx, cfg, logger, grpcServer)
	if err != nil {
		logger.Error(ctx, err, "Failed to create REST gateway", 500)
		return nil, fmt.Errorf("failed to create REST gateway: %w", err)
	}

	return &Server{
		logger:     logger,
		config:     cfg,
		db:         db,
		service:    svc,
		grpcServer: grpcServer,
		restServer: restServer,
		grpcLis:    grpcLis,
		restLis:    restLis,
	}, nil
}

// createRESTGateway creates the REST gateway server
func createRESTGateway(ctx context.Context, cfg *config.Config, logger *zlog.Logger, grpcServer *grpc.Server) (*http.Server, net.Listener, error) {
	// Create REST listener
	restLis, err := net.Listen("tcp", ":"+cfg.RestGatewayPort)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create REST listener: %w", err)
	}

	// Create gRPC-Gateway mux with custom options
	gwMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames:   true,
				EmitUnpopulated: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
			// Custom error handler to provide better error messages
			logger.Error(ctx, err, "REST gateway error", 500, map[string]any{
				"path":   r.URL.Path,
				"method": r.Method,
			})
			
			// Check if it's a gRPC connection error
			if status.Code(err) == codes.Unavailable || 
			   status.Code(err) == codes.DeadlineExceeded ||
			   err.Error() == "grpc: the client connection is closing" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error":"Service temporarily unavailable","code":"SERVICE_UNAVAILABLE"}`))
				return
			}
			
			// Use default error handler for other errors
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
		}),
	)

	// Add direct health check endpoint as fallback
	http.HandleFunc("/v1/health/direct", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add a simple health endpoint that doesn't depend on gRPC
	http.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add a root health endpoint for basic connectivity testing
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"auth-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Register REST handlers
	// Use localhost for internal container communication
	grpcAddr := "localhost:" + cfg.AuthServicePort
	
	// Create a shared gRPC connection for better connection management
	var dialOptions []grpc.DialOption
	
	// Configure TLS if enabled
	if cfg.TLSEnabled {
		tlsConfig, err := createTLSConfig(cfg)
		if err != nil {
			logger.Error(ctx, err, "Failed to create TLS config for gRPC connection", 500)
			return nil, nil, fmt.Errorf("failed to create TLS config for gRPC connection: %w", err)
		}
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	}
	
	// Add other connection options
	dialOptions = append(dialOptions,
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 10 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  1 * time.Second,
				Multiplier: 1.6,
				Jitter:     0.2,
				MaxDelay:   120 * time.Second,
			},
		}),
		// Add keepalive settings to maintain connection
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		// Disable retry to prevent connection issues
		grpc.WithDisableRetry(),
	)
	
	sharedConn, err := grpc.Dial(grpcAddr, dialOptions...)
	if err != nil {
		logger.Error(ctx, err, "Failed to create shared gRPC connection", 500)
		return nil, nil, fmt.Errorf("failed to create shared gRPC connection: %w", err)
	}

	// Register AuthService REST handler using shared connection
	if err := proto.RegisterAuthServiceHandlerClient(ctx, gwMux, proto.NewAuthServiceClient(sharedConn)); err != nil {
		logger.Error(ctx, err, "Failed to register AuthService REST handler", 500)
		sharedConn.Close()
		return nil, nil, fmt.Errorf("failed to register AuthService REST handler: %w", err)
	}

	// Register Health service REST handler using shared connection
	if err := proto.RegisterHealthHandlerClient(ctx, gwMux, proto.NewHealthClient(sharedConn)); err != nil {
		logger.Error(ctx, err, "Failed to register Health REST handler", 500)
		sharedConn.Close()
		return nil, nil, fmt.Errorf("failed to register Health REST handler: %w", err)
	}

	// Create HTTP server with proper timeout configurations
	restServer := &http.Server{
		Handler:           createRESTMiddleware(gwMux, logger, cfg),
		Addr:              restLis.Addr().String(),
		ReadTimeout:       time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return restServer, restLis, nil
}

// isAllowedOrigin checks if an origin is in the allowed origins list
func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// createTLSConfig creates a TLS configuration for the server
func createTLSConfig(cfg *config.Config) (*tls.Config, error) {
	// Load certificate and private key
	cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:  cfg.MinTLSVersion,
		MaxVersion:  cfg.MaxTLSVersion,
		
		// Security best practices
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		
		// Prefer server cipher suites
		PreferServerCipherSuites: true,
		
		// Curve preferences for ECDHE
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
		},
	}

	return tlsConfig, nil
}

// createRESTMiddleware creates middleware for the REST gateway
func createRESTMiddleware(handler http.Handler, logger *zlog.Logger, cfg *config.Config) http.Handler {
	// Apply security headers middleware first
	handler = CreateSecurityHeadersMiddleware(cfg)(handler)
	
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers with proper origin validation
		origin := r.Header.Get("Origin")
		if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Add request logging
		logger.Info(r.Context(), "REST request", map[string]any{
			"method":     r.Method,
			"path":       r.URL.Path,
			"remote":     r.RemoteAddr,
			"user_agent": r.UserAgent(),
		})

		// Add response logging
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
		handler.ServeHTTP(responseWriter, r)

		// Log response
		logger.Info(r.Context(), "REST response", map[string]any{
			"method": r.Method,
			"path":   r.URL.Path,
			"status": responseWriter.statusCode,
		})
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// Start runs both gRPC and REST servers
func (s *Server) Start(ctx context.Context) error {
	s.logger.Info(ctx, "Starting gRPC service", map[string]any{
		"port": s.config.AuthServicePort,
	})

	// Start gRPC server in a goroutine
	go func() {
		if err := s.grpcServer.Serve(s.grpcLis); err != nil {
			s.logger.Error(ctx, err, "gRPC server failed to start", 500)
			os.Exit(1)
		}
	}()

	// Wait for gRPC server to be ready by checking if it's listening
	ready := make(chan bool, 1)
	go func() {
		maxAttempts := 50 // 5 seconds with 100ms intervals
		attempts := 0
		for attempts < maxAttempts {
			select {
			case <-ctx.Done():
				return
			default:
				// Try to connect to the gRPC server to check if it's ready
				conn, err := grpc.Dial("localhost:"+s.config.AuthServicePort, 
					grpc.WithInsecure(),
					grpc.WithBlock(),
					grpc.WithTimeout(100*time.Millisecond))
				if err == nil {
					conn.Close()
					ready <- true
					return
				}
				attempts++
				time.Sleep(100 * time.Millisecond)
			}
		}
		// If we reach here, the server didn't become ready
		s.logger.Warn(ctx, "gRPC server readiness check timed out", map[string]any{
			"max_attempts": maxAttempts,
			"timeout_ms":   maxAttempts * 100,
		})
	}()

	// Wait for gRPC server to be ready or timeout
	select {
	case <-ready:
		s.logger.Info(ctx, "gRPC server is ready")
	case <-time.After(10 * time.Second):
		s.logger.Warn(ctx, "gRPC server readiness check timed out, proceeding anyway")
	case <-ctx.Done():
		return ctx.Err()
	}

	s.logger.Info(ctx, "Starting REST gateway", map[string]any{
		"port": s.config.RestGatewayPort,
	})

	// Start REST server in a goroutine
	go func() {
		if err := s.restServer.Serve(s.restLis); err != nil && err != http.ErrServerClosed {
			s.logger.Error(ctx, err, "REST server failed to start", 500)
			os.Exit(1)
		}
	}()

	// Wait a moment for REST server to start
	time.Sleep(100 * time.Millisecond)

	s.logger.Info(ctx, "All servers started successfully", map[string]any{
		"grpc_port": s.config.AuthServicePort,
		"rest_port": s.config.RestGatewayPort,
	})

	return nil
}

// Shutdown gracefully shuts down both servers
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info(ctx, "Shutting down servers")

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	// Graceful stop the gRPC server
	s.grpcServer.GracefulStop()

	// Shutdown the REST server
	if err := s.restServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error(shutdownCtx, err, "Failed to shutdown REST server", 500)
	}

	// Close database connection
	if err := s.db.Close(shutdownCtx); err != nil {
		s.logger.Error(shutdownCtx, err, "Failed to close database", 500)
		return err
	}

	s.logger.Info(ctx, "Server shutdown completed")
	return nil
}

// Run starts both servers and handles graceful shutdown
func (s *Server) Run(ctx context.Context) error {
	// Start the servers
	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("failed to start servers: %v", err)
	}

	// Set up signal handling with buffered channel
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	s.logger.Info(ctx, "Received shutdown signal")

	// Perform graceful shutdown
	if err := s.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown servers: %v", err)
	}

	return nil
}
