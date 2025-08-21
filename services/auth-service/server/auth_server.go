package server

import (
	"context"
	"time"

	"auth-service/models"
	"auth-service/proto"
	"auth-service/services"
	zlog "packages/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthServer implements the AuthService gRPC interface
type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	service *services.Service
	logger  *zlog.Logger
}

// NewAuthServer creates a new auth server instance
func NewAuthServer(svc *services.Service, logger *zlog.Logger) *AuthServer {
	return &AuthServer{
		service: svc,
		logger:  logger,
	}
}

// SignUp handles user registration
func (s *AuthServer) SignUp(ctx context.Context, req *proto.UserCreateRequest) (*proto.AuthResponse, error) {
	s.logger.Info(ctx, "Processing SignUp request", map[string]any{
		"email": req.Email,
	})

	// Convert protobuf request to internal model
	userReq := &models.UserCreateRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call service
	user, err := s.service.Auth.SignUp(ctx, userReq)
	if err != nil {
		s.logger.Error(ctx, err, "SignUp failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "signup failed: %v", err)
	}

	// Generate tokens for the user using secrets from config
	accessToken, refreshToken, err := s.service.Auth.GenerateTokens(ctx, user, s.service.Config.JWTAccessTokenSecret, s.service.Config.JWTRefreshTokenSecret)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to generate tokens", 500)
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	// Create user token model
	userToken := &models.UserToken{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now(),
	}

	// Convert to protobuf response
	response := &proto.AuthResponse{
		User:   convertUserToProto(user),
		Tokens: convertUserTokenToProto(userToken),
	}

	s.logger.Info(ctx, "SignUp completed successfully", map[string]any{
		"user_id": user.ID.String(),
	})

	return response, nil
}

// SignIn handles user authentication
func (s *AuthServer) SignIn(ctx context.Context, req *proto.Credentials) (*proto.AuthResponse, error) {
	s.logger.Info(ctx, "Processing SignIn request", map[string]any{
		"email": req.Email,
	})

	// Convert protobuf request to internal model
	creds := &models.Credentials{
		Email:    req.Email,
		Password: req.Password,
	}

	// Call service with JWT secrets
	user, accessToken, refreshToken, err := s.service.Auth.SignIn(ctx, creds, s.service.Config.JWTAccessTokenSecret, s.service.Config.JWTRefreshTokenSecret)
	if err != nil {
		s.logger.Error(ctx, err, "SignIn failed", 401)
		return nil, status.Errorf(codes.Unauthenticated, "signin failed: %v", err)
	}

	// Create user token model
	userToken := &models.UserToken{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    time.Now(),
	}

	// Convert to protobuf response
	response := &proto.AuthResponse{
		User:   convertUserToProto(user),
		Tokens: convertUserTokenToProto(userToken),
	}

	s.logger.Info(ctx, "SignIn completed successfully", map[string]any{
		"user_id": user.ID.String(),
	})

	return response, nil
}

// SignOut handles user sign out
func (s *AuthServer) SignOut(ctx context.Context, req *proto.SignOutRequest) (*proto.Empty, error) {
	s.logger.Info(ctx, "Processing SignOut request")

	// Call service
	err := s.service.Auth.Signout(ctx, req.AccessToken)
	if err != nil {
		s.logger.Error(ctx, err, "SignOut failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "signout failed: %v", err)
	}

	s.logger.Info(ctx, "SignOut completed successfully")

	return &proto.Empty{}, nil
}

// RefreshToken handles token refresh
func (s *AuthServer) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.TokenResponse, error) {
	s.logger.Info(ctx, "Processing RefreshToken request")

	// Call service with JWT secrets
	tokens, err := s.service.Auth.RefreshToken(ctx, req.RefreshToken, s.service.Config.JWTAccessTokenSecret, s.service.Config.JWTRefreshTokenSecret)
	if err != nil {
		s.logger.Error(ctx, err, "RefreshToken failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "token refresh failed: %v", err)
	}

	// Convert to protobuf response
	response := &proto.TokenResponse{
		Tokens: convertUserTokenToProto(tokens),
	}

	s.logger.Info(ctx, "RefreshToken completed successfully")

	return response, nil
}

// RevokeToken handles token revocation
func (s *AuthServer) RevokeToken(ctx context.Context, req *proto.RevokeTokenRequest) (*proto.Empty, error) {
	s.logger.Info(ctx, "Processing RevokeToken request")

	// Call service
	err := s.service.Auth.RevokeToken(ctx, req.AccessToken)
	if err != nil {
		s.logger.Error(ctx, err, "RevokeToken failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "token revocation failed: %v", err)
	}

	s.logger.Info(ctx, "RevokeToken completed successfully")

	return &proto.Empty{}, nil
}

// ListUsers handles user listing
func (s *AuthServer) ListUsers(ctx context.Context, req *proto.ListUsersRequest) (*proto.ListUsersResponse, error) {
	s.logger.Info(ctx, "Processing ListUsers request", map[string]any{
		"page":  req.Page,
		"limit": req.Limit,
	})

	// Set default values for pagination
	page := int(req.Page)
	limit := int(req.Limit)

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// Call service
	users, total, err := s.service.User.GetAllUsers(ctx, page, limit)
	if err != nil {
		s.logger.Error(ctx, err, "ListUsers failed", 500)
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// Convert to protobuf response
	protoUsers := make([]*proto.User, len(users))
	for i, user := range users {
		protoUsers[i] = convertUserToProto(&user)
	}

	response := &proto.ListUsersResponse{
		Users: protoUsers,
		Total: int32(total),
		Page:  int32(page),
		Limit: int32(limit),
	}

	s.logger.Info(ctx, "ListUsers completed successfully", map[string]any{
		"count": len(users),
	})

	return response, nil
}

// ValidateToken handles token validation
func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	s.logger.Info(ctx, "Processing ValidateToken request")

	// Call service with JWT secret
	user, err := s.service.Auth.ValidateToken(ctx, req.Token, s.service.Config.JWTAccessTokenSecret)
	if err != nil {
		s.logger.Error(ctx, err, "ValidateToken failed", 401)
		return &proto.ValidateTokenResponse{
			Valid:        false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &proto.ValidateTokenResponse{
		UserId: user.ID.String(),
		Valid:  true,
	}, nil
}

// Helper functions to convert between internal models and protobuf messages

func convertUserToProto(user *models.User) *proto.User {
	return &proto.User{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

func convertUserTokenToProto(token *models.UserToken) *proto.UserToken {
	return &proto.UserToken{
		Id:               token.ID.String(),
		UserId:           token.UserID.String(),
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		AccessExpiresAt:  timestamppb.New(token.AccessExpiresAt),
		RefreshExpiresAt: timestamppb.New(token.RefreshExpiresAt),
		IsRevoked:        token.IsRevoked,
		CreatedAt:        timestamppb.New(token.CreatedAt),
	}
}
