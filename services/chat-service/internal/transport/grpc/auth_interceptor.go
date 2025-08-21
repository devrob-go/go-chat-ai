package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"auth-service/proto"
	"chat-service/configs"
	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor handles authentication by calling the auth service
type AuthInterceptor struct {
	logger   *zlog.Logger
	config   *configs.Config
	authConn *grpc.ClientConn
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(logger *zlog.Logger, config *configs.Config) (*AuthInterceptor, error) {
	// Load client certificates for mTLS only if TLS is enabled
	var tlsConfig *tls.Config
	if config.AuthServiceTLS && config.TLSEnabled {
		cert, err := tls.LoadX509KeyPair(config.AuthServiceCertFile, config.AuthServiceKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificates: %w", err)
		}

		// Load CA certificate
		caCert, err := ioutil.ReadFile(config.AuthServiceCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   config.AuthServiceHost,
		}
	}

	// Create gRPC connection to auth service
	var creds credentials.TransportCredentials
	if config.AuthServiceTLS && config.TLSEnabled {
		creds = credentials.NewTLS(tlsConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	authConn, err := grpc.Dial(
		fmt.Sprintf("%s:%s", config.AuthServiceHost, config.AuthServicePort),
		grpc.WithTransportCredentials(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	return &AuthInterceptor{
		logger:   logger,
		config:   config,
		authConn: authConn,
	}, nil
}

// UnaryAuthInterceptor intercepts unary gRPC calls for authentication
func (i *AuthInterceptor) UnaryAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Skip authentication for health checks
		if info.FullMethod == "/health.Health/Check" || info.FullMethod == "/health.Health/Watch" {
			return handler(ctx, req)
		}

		// Extract token from metadata
		token, err := i.extractToken(ctx)
		if err != nil {
			i.logger.Warn(ctx, "Failed to extract token", map[string]interface{}{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Validate token with auth service
		userID, err := i.validateToken(ctx, token)
		if err != nil {
			i.logger.Warn(ctx, "Token validation failed", map[string]interface{}{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return nil, status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
		}

		// Add user ID to context
		ctx = context.WithValue(ctx, "user_id", userID)

		i.logger.Debug(ctx, "Authentication successful", map[string]interface{}{
			"method":  info.FullMethod,
			"user_id": userID,
		})

		return handler(ctx, req)
	}
}

// StreamAuthInterceptor intercepts streaming gRPC calls for authentication
func (i *AuthInterceptor) StreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Skip authentication for health checks
		if info.FullMethod == "/health.Health/Check" || info.FullMethod == "/health.Health/Watch" {
			return handler(srv, stream)
		}

		ctx := stream.Context()

		// Extract token from metadata
		token, err := i.extractToken(ctx)
		if err != nil {
			i.logger.Warn(ctx, "Failed to extract token", map[string]interface{}{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Validate token with auth service
		userID, err := i.validateToken(ctx, token)
		if err != nil {
			i.logger.Warn(ctx, "Token validation failed", map[string]interface{}{
				"method": info.FullMethod,
				"error":  err.Error(),
			})
			return status.Errorf(codes.Unauthenticated, "token validation failed: %v", err)
		}

		// Create new context with user ID
		newCtx := context.WithValue(ctx, "user_id", userID)

		// Create wrapped stream with new context
		wrappedStream := &wrappedServerStream{
			ServerStream: stream,
			ctx:          newCtx,
		}

		i.logger.Debug(ctx, "Authentication successful for stream", map[string]interface{}{
			"method":  info.FullMethod,
			"user_id": userID,
		})

		return handler(srv, wrappedStream)
	}
}

// extractToken extracts the JWT token from the gRPC metadata
func (i *AuthInterceptor) extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata found")
	}

	authHeader := md.Get("authorization")
	if len(authHeader) == 0 {
		return "", fmt.Errorf("no authorization header found")
	}

	token := authHeader[0]
	if len(token) < 7 || token[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return token[7:], nil
}

// validateToken validates the token with the auth service
func (i *AuthInterceptor) validateToken(ctx context.Context, token string) (string, error) {
	// Create auth service client
	authClient := proto.NewAuthServiceClient(i.authConn)

	// Call the auth service to validate the token
	resp, err := authClient.ValidateToken(ctx, &proto.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return "", fmt.Errorf("auth service error: %w", err)
	}

	// Check if token is valid
	if !resp.Valid {
		return "", fmt.Errorf("token validation failed: %s", resp.ErrorMessage)
	}

	return resp.UserId, nil
}

// Close closes the auth service connection
func (i *AuthInterceptor) Close() error {
	if i.authConn != nil {
		return i.authConn.Close()
	}
	return nil
}

// wrappedServerStream wraps the gRPC server stream to provide a custom context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
