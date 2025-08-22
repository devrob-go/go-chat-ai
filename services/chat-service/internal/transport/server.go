package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	authproto "api/auth/v1/proto"
	"chat-service/configs"
	"chat-service/internal/domain"
	"chat-service/internal/services/chat"
	"chat-service/internal/services/openai"
	grpchandler "chat-service/internal/transport/grpc"
	chatproto "chat-service/proto"
	"chat-service/storage"
	zlog "packages/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

const (
	// DefaultShutdownTimeout is the default timeout for graceful shutdown
	DefaultShutdownTimeout = 5 * time.Second
)

// createRESTGateway creates the REST gateway server
func createRESTGateway(ctx context.Context, cfg *configs.Config, logger *zlog.Logger, grpcServer *grpc.Server, chatService chat.Service) (*http.Server, net.Listener, error) {
	// Create REST listener
	restLis, err := net.Listen("tcp", ":"+cfg.RestGatewayPort)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create REST listener: %w", err)
	}

	// Create a simple HTTP mux for now
	mux := http.NewServeMux()

	// Add direct health check endpoint as fallback
	mux.HandleFunc("/v1/health/direct", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"chat-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add a simple health endpoint that doesn't depend on gRPC
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"chat-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Add a root health endpoint for basic connectivity testing
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"SERVING","service":"chat-service","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Chat endpoints
	mux.HandleFunc("/v1/chat/message", func(w http.ResponseWriter, r *http.Request) {
		handleSendMessage(w, r, chatService, logger, cfg)
	})

	mux.HandleFunc("/v1/chat/ai", func(w http.ResponseWriter, r *http.Request) {
		handleChatWithAI(w, r, chatService, logger, cfg)
	})

	mux.HandleFunc("/v1/chat/conversations", func(w http.ResponseWriter, r *http.Request) {
		handleConversations(w, r, chatService, logger, cfg)
	})

	mux.HandleFunc("/v1/chat/history/", func(w http.ResponseWriter, r *http.Request) {
		handleGetHistory(w, r, chatService, logger, cfg)
	})

	// Create HTTP server with proper timeout configurations
	restServer := &http.Server{
		Handler:           mux,
		Addr:              restLis.Addr().String(),
		ReadTimeout:       time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout:      time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
	}

	return restServer, restLis, nil
}

// extractUserIDFromToken extracts user ID from JWT token in REST requests
func extractUserIDFromToken(r *http.Request, config *configs.Config) (string, error) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header found")
	}

	// Check Bearer token format
	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := authHeader[7:]

	// Create a context for the gRPC call
	ctx := r.Context()

	// Create gRPC connection to auth service
	var authConn *grpc.ClientConn
	var err error

	if config.AuthServiceTLS && config.TLSEnabled {
		// Load client certificates for mTLS
		cert, err := tls.LoadX509KeyPair(config.AuthServiceCertFile, config.AuthServiceKeyFile)
		if err != nil {
			return "", fmt.Errorf("failed to load client certificates: %w", err)
		}

		// Load CA certificate
		caCert, err := ioutil.ReadFile(config.AuthServiceCAFile)
		if err != nil {
			return "", fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return "", fmt.Errorf("failed to append CA certificate")
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
			ServerName:   config.AuthServiceHost,
		}

		creds := credentials.NewTLS(tlsConfig)
		authConn, err = grpc.Dial(
			fmt.Sprintf("%s:%s", config.AuthServiceHost, config.AuthServicePort),
			grpc.WithTransportCredentials(creds),
		)
	} else {
		authConn, err = grpc.Dial(
			fmt.Sprintf("%s:%s", config.AuthServiceHost, config.AuthServicePort),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
	}

	if err != nil {
		return "", fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer authConn.Close()

	// Create auth service client and validate token
	authClient := authproto.NewAuthServiceClient(authConn)
	resp, err := authClient.ValidateToken(ctx, &authproto.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		return "", fmt.Errorf("auth service error: %w", err)
	}

	if !resp.Valid {
		return "", fmt.Errorf("token validation failed: %s", resp.ErrorMessage)
	}

	return resp.UserId, nil
}

// REST endpoint handlers

// handleSendMessage handles POST /v1/chat/message
func handleSendMessage(w http.ResponseWriter, r *http.Request, chatService chat.Service, logger *zlog.Logger, config *configs.Config) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(domain.NewErrorResponse("METHOD_NOT_ALLOWED", "Method not allowed", "405"))
		return
	}

	// Extract user ID from JWT token
	userID, err := extractUserIDFromToken(r, config)
	if err != nil {
		logger.Error(r.Context(), err, "Failed to extract user ID from token", 401)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(domain.NewErrorResponse("UNAUTHORIZED", "Unauthorized", "401"))
		return
	}

	// Parse request body
	var req struct {
		Message        string `json:"message"`
		ConversationID string `json:"conversation_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.NewErrorResponse("INVALID_REQUEST", "Invalid request body", "400"))
		return
	}

	// Validate required fields
	if req.Message == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.NewErrorResponse("VALIDATION_ERROR", "message is required", "400"))
		return
	}

	// Create domain request
	domainReq := &domain.ChatRequest{
		UserID:         userID,
		Message:        req.Message,
		ConversationID: req.ConversationID,
	}

	// Validate the request
	if err := domainReq.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(domain.NewErrorResponseWithDetails("VALIDATION_ERROR", "Validation error", "400", map[string]string{
			"details": err.Error(),
		}))
		return
	}

	// Call chat service
	ctx := r.Context()
	response, err := chatService.SendMessage(ctx, domainReq)
	if err != nil {
		logger.Error(ctx, err, "Failed to send message", 500)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(domain.NewErrorResponse("INTERNAL_ERROR", "Internal server error", "500"))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message": map[string]any{
			"id":         response.Message.ID,
			"user_id":    response.Message.UserID,
			"content":    response.Message.Content,
			"role":       response.Message.Role,
			"created_at": response.Message.CreatedAt,
		},
		"conversation_id": response.ConversationID,
		"is_ai_response":  response.IsAIResponse,
	})
}

// handleChatWithAI handles POST /v1/chat/ai
func handleChatWithAI(w http.ResponseWriter, r *http.Request, chatService chat.Service, logger *zlog.Logger, config *configs.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from JWT token
	userID, err := extractUserIDFromToken(r, config)
	if err != nil {
		logger.Error(r.Context(), err, "Failed to extract user ID from token", 401)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Message        string  `json:"message"`
		ConversationID string  `json:"conversation_id,omitempty"`
		Model          string  `json:"model,omitempty"`
		Temperature    float64 `json:"temperature,omitempty"`
		MaxTokens      int     `json:"max_tokens,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Message == "" {
		http.Error(w, "message is required", http.StatusBadRequest)
		return
	}

	// Validate UUIDs
	if err := domain.ValidateUUID(userID); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}
	if req.ConversationID != "" {
		if err := domain.ValidateUUID(req.ConversationID); err != nil {
			http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Set defaults
	if req.Model == "" {
		req.Model = "gpt-3.5-turbo"
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 1000
	}

	// Call chat service
	ctx := r.Context()
	response, err := chatService.ChatWithAI(ctx, userID, req.Message, req.ConversationID, req.Model, req.Temperature, req.MaxTokens)
	if err != nil {
		logger.Error(ctx, err, "Failed to chat with AI", 500)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"ai_message":      response.Message.Content,
		"conversation_id": response.ConversationID,
		"model_used":      req.Model,
		"tokens_used":     0, // Would come from OpenAI response
		"created_at":      response.Message.CreatedAt,
	})
}

// handleConversations handles GET/POST /v1/chat/conversations
func handleConversations(w http.ResponseWriter, r *http.Request, chatService chat.Service, logger *zlog.Logger, config *configs.Config) {
	switch r.Method {
	case http.MethodGet:
		handleListConversations(w, r, chatService, logger, config)
	case http.MethodPost:
		handleCreateConversation(w, r, chatService, logger, config)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListConversations handles GET /v1/chat/conversations
func handleListConversations(w http.ResponseWriter, r *http.Request, chatService chat.Service, logger *zlog.Logger, config *configs.Config) {
	// Extract user ID from JWT token
	userID, err := extractUserIDFromToken(r, config)
	if err != nil {
		logger.Error(r.Context(), err, "Failed to extract user ID from token", 401)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 10 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Create domain request
	domainReq := &domain.ListConversationsRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Validate the request
	if err := domainReq.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Call chat service
	ctx := r.Context()
	response, err := chatService.ListConversations(ctx, domainReq)
	if err != nil {
		logger.Error(ctx, err, "Failed to list conversations", 500)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	conversations := make([]map[string]any, len(response.Conversations))
	for i, conv := range response.Conversations {
		conversations[i] = map[string]any{
			"id":         conv.ID,
			"user_id":    conv.UserID,
			"title":      conv.Title,
			"created_at": conv.CreatedAt,
			"updated_at": conv.UpdatedAt,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"conversations": conversations,
		"total":         response.Total,
	})
}

// handleCreateConversation handles POST /v1/chat/conversations
func handleCreateConversation(w http.ResponseWriter, r *http.Request, chatService chat.Service, logger *zlog.Logger, config *configs.Config) {
	// Extract user ID from JWT token
	userID, err := extractUserIDFromToken(r, config)
	if err != nil {
		logger.Error(r.Context(), err, "Failed to extract user ID from token", 401)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	// Validate UUID
	if err := domain.ValidateUUID(userID); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Call chat service
	ctx := r.Context()
	conversation, err := chatService.CreateConversation(ctx, userID, req.Title)
	if err != nil {
		logger.Error(ctx, err, "Failed to create conversation", 500)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"id":         conversation.ID,
		"user_id":    conversation.UserID,
		"title":      conversation.Title,
		"created_at": conversation.CreatedAt,
		"updated_at": conversation.UpdatedAt,
	})
}

// handleGetHistory handles GET /v1/chat/history/{conversation_id}
func handleGetHistory(w http.ResponseWriter, r *http.Request, chatService chat.Service, logger *zlog.Logger, config *configs.Config) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract conversation ID from URL path
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/chat/history/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "conversation_id is required in URL path", http.StatusBadRequest)
		return
	}
	conversationID := pathParts[0]

	// Validate conversation ID UUID
	if err := domain.ValidateUUID(conversationID); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Extract user ID from JWT token
	userID, err := extractUserIDFromToken(r, config)
	if err != nil {
		logger.Error(r.Context(), err, "Failed to extract user ID from token", 401)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit := 50 // default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Create domain request
	domainReq := &domain.GetHistoryRequest{
		UserID:         userID,
		ConversationID: conversationID,
		Limit:          limit,
		Offset:         offset,
	}

	// Validate the request
	if err := domainReq.Validate(); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}

	// Call chat service
	ctx := r.Context()
	response, err := chatService.GetHistory(ctx, domainReq)
	if err != nil {
		logger.Error(ctx, err, "Failed to get chat history", 500)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	messages := make([]map[string]any, len(response.Messages))
	for i, msg := range response.Messages {
		messages[i] = map[string]any{
			"id":         msg.ID,
			"user_id":    msg.UserID,
			"content":    msg.Content,
			"role":       msg.Role,
			"created_at": msg.CreatedAt,
			"updated_at": msg.UpdatedAt,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"messages":        messages,
		"total":           response.Total,
		"conversation_id": response.ConversationID,
	})
}

// Server holds the gRPC server and its dependencies
type Server struct {
	logger          *zlog.Logger
	config          *configs.Config
	grpcServer      *grpc.Server
	grpcLis         net.Listener
	restServer      *http.Server
	restLis         net.Listener
	authInterceptor *grpchandler.AuthInterceptor
	db              *storage.DB
}

// NewServer initializes the gRPC server with its dependencies
func NewServer(ctx context.Context) (*Server, error) {
	// Load configuration
	cfg, err := configs.LoadConfig()
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

	// Initialize OpenAI client
	logger.Info(ctx, "Initializing OpenAI client")
	openaiClient := openai.NewClient(cfg, logger)

	// Initialize storage
	logger.Info(ctx, "Initializing database storage")
	db, err := storage.InitDB(ctx, cfg, logger)
	if err != nil {
		logger.Error(ctx, err, "Failed to initialize database storage", 500)
		return nil, fmt.Errorf("failed to initialize database storage: %w", err)
	}

	// Initialize chat service
	logger.Info(ctx, "Creating chat service")
	chatService := chat.NewService(openaiClient, logger, cfg, db)

	// Initialize auth interceptor
	logger.Info(ctx, "Initializing auth interceptor")
	authInterceptor, err := grpchandler.NewAuthInterceptor(logger, cfg)
	if err != nil {
		logger.Error(ctx, err, "Failed to initialize auth interceptor", 500)
		return nil, fmt.Errorf("failed to initialize auth interceptor: %w", err)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			authInterceptor.UnaryAuthInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			authInterceptor.StreamAuthInterceptor(),
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			MaxConnectionAge:  10 * time.Minute,
			Time:              2 * time.Minute,
			Timeout:           20 * time.Second,
		}),
	)

	// Register services
	logger.Info(ctx, "Registering gRPC services")
	chatproto.RegisterChatServiceServer(grpcServer, grpchandler.NewChatHandler(chatService, logger))

	// Enable reflection for development
	if cfg.Environment == configs.DEVELOPMENT_ENV {
		reflection.Register(grpcServer)
		logger.Info(ctx, "gRPC reflection enabled for development")
	}

	// Create gRPC listener
	var grpcLis net.Listener
	if cfg.TLSEnabled {
		// Load TLS certificates
		cert, err := tls.LoadX509KeyPair(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			logger.Error(ctx, err, "Failed to load TLS certificates", 500)
			return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   cfg.MinTLSVersion,
			MaxVersion:   cfg.MaxTLSVersion,
		}

		grpcLis, err = tls.Listen("tcp", ":"+cfg.ChatServicePort, tlsConfig)
		if err != nil {
			logger.Error(ctx, err, "Failed to create TLS listener", 500)
			return nil, fmt.Errorf("failed to create TLS listener: %w", err)
		}

		logger.Info(ctx, "TLS enabled for gRPC server", map[string]any{
			"port": cfg.ChatServicePort,
		})
	} else {
		grpcLis, err = net.Listen("tcp", ":"+cfg.ChatServicePort)
		if err != nil {
			logger.Error(ctx, err, "Failed to create listener", 500)
			return nil, fmt.Errorf("failed to create listener: %w", err)
		}

		logger.Info(ctx, "TLS disabled for gRPC server", map[string]any{
			"port": cfg.ChatServicePort,
		})
	}

	// Create REST gateway
	restServer, restLis, err := createRESTGateway(ctx, cfg, logger, grpcServer, chatService)
	if err != nil {
		logger.Error(ctx, err, "Failed to create REST gateway", 500)
		return nil, fmt.Errorf("failed to create REST gateway: %w", err)
	}

	return &Server{
		logger:          logger,
		config:          cfg,
		grpcServer:      grpcServer,
		grpcLis:         grpcLis,
		restServer:      restServer,
		restLis:         restLis,
		authInterceptor: authInterceptor,
		db:              db,
	}, nil
}

// Run starts the server and waits for shutdown signal
func (s *Server) Run(ctx context.Context) error {
	// Start gRPC server in a goroutine
	go func() {
		s.logger.Info(ctx, "Starting gRPC server", map[string]any{
			"port": s.config.ChatServicePort,
		})

		if err := s.grpcServer.Serve(s.grpcLis); err != nil {
			s.logger.Error(ctx, err, "Failed to serve gRPC", 500)
		}
	}()

	// Start REST server in a goroutine
	go func() {
		s.logger.Info(ctx, "Starting REST server", map[string]any{
			"port": s.config.RestGatewayPort,
		})

		if err := s.restServer.Serve(s.restLis); err != nil {
			s.logger.Error(ctx, err, "Failed to serve REST", 500)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info(ctx, "Shutting down server...")

	// Graceful shutdown
	return s.Shutdown(ctx)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(ctx, DefaultShutdownTimeout)
	defer cancel()

	// Close auth interceptor
	if s.authInterceptor != nil {
		if err := s.authInterceptor.Close(); err != nil {
			s.logger.Warn(ctx, "Failed to close auth interceptor", map[string]any{
				"error": err.Error(),
			})
		}
	}

	// Close database connection
	if s.db != nil {
		if err := s.db.Close(ctx); err != nil {
			s.logger.Warn(ctx, "Failed to close database connection", map[string]any{
				"error": err.Error(),
			})
		}
	}

	// Gracefully stop the gRPC server
	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	// Wait for either graceful stop or timeout
	select {
	case <-done:
		s.logger.Info(ctx, "Server stopped gracefully")
	case <-ctx.Done():
		s.logger.Warn(ctx, "Server shutdown timed out, forcing stop")
		s.grpcServer.Stop()
	}

	// Close listeners
	if s.grpcLis != nil {
		if err := s.grpcLis.Close(); err != nil {
			s.logger.Warn(ctx, "Failed to close gRPC listener", map[string]any{
				"error": err.Error(),
			})
		}
	}

	if s.restLis != nil {
		if err := s.restLis.Close(); err != nil {
			s.logger.Warn(ctx, "Failed to close REST listener", map[string]any{
				"error": err.Error(),
			})
		}
	}

	s.logger.Info(ctx, "Server shutdown completed")
	return nil
}
