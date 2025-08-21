package grpc

import (
	"context"

	"auth-service/internal/domain"
	"auth-service/proto"
	zlog "packages/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthHandler handles gRPC authentication requests
type AuthHandler struct {
	proto.UnimplementedAuthServiceServer
	authService          domain.AuthService
	userService          domain.UserService
	logger               *zlog.Logger
	jwtAccessTokenSecret string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService domain.AuthService, userService domain.UserService, logger *zlog.Logger, jwtAccessTokenSecret string) *AuthHandler {
	return &AuthHandler{
		authService:          authService,
		userService:          userService,
		logger:               logger,
		jwtAccessTokenSecret: jwtAccessTokenSecret,
	}
}

// SignUp handles user registration
func (h *AuthHandler) SignUp(ctx context.Context, req *proto.UserCreateRequest) (*proto.AuthResponse, error) {
	h.logger.Info(ctx, "Processing SignUp request", map[string]any{
		"email": req.Email,
	})

	// Call domain service
	response, err := h.authService.SignUp(req.Name, req.Email, req.Password)
	if err != nil {
		h.logger.Error(ctx, err, "SignUp failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "signup failed: %v", err)
	}

	// Convert domain response to proto
	return &proto.AuthResponse{
		User: &proto.User{
			Id:        response.User.ID.String(),
			Name:      response.User.Name,
			Email:     response.User.Email,
			CreatedAt: timestamppb.New(response.User.CreatedAt),
			UpdatedAt: timestamppb.New(response.User.UpdatedAt),
		},
		Tokens: &proto.UserToken{
			Id:               response.Tokens.ID.String(),
			UserId:           response.Tokens.UserID.String(),
			AccessToken:      response.Tokens.AccessToken,
			RefreshToken:     response.Tokens.RefreshToken,
			AccessExpiresAt:  timestamppb.New(response.Tokens.AccessExpiresAt),
			RefreshExpiresAt: timestamppb.New(response.Tokens.RefreshExpiresAt),
			IsRevoked:        response.Tokens.IsRevoked,
			CreatedAt:        timestamppb.New(response.Tokens.CreatedAt),
		},
	}, nil
}

// SignIn handles user authentication
func (h *AuthHandler) SignIn(ctx context.Context, req *proto.Credentials) (*proto.AuthResponse, error) {
	h.logger.Info(ctx, "Processing SignIn request", map[string]any{
		"email": req.Email,
	})

	// Call domain service
	response, err := h.authService.SignIn(req.Email, req.Password)
	if err != nil {
		h.logger.Error(ctx, err, "SignIn failed", 401)
		return nil, status.Errorf(codes.Unauthenticated, "signin failed: %v", err)
	}

	// Convert domain response to proto
	return &proto.AuthResponse{
		User: &proto.User{
			Id:        response.User.ID.String(),
			Name:      response.User.Name,
			Email:     response.User.Email,
			CreatedAt: timestamppb.New(response.User.CreatedAt),
			UpdatedAt: timestamppb.New(response.User.UpdatedAt),
		},
		Tokens: &proto.UserToken{
			Id:               response.Tokens.ID.String(),
			UserId:           response.Tokens.UserID.String(),
			AccessToken:      response.Tokens.AccessToken,
			RefreshToken:     response.Tokens.RefreshToken,
			AccessExpiresAt:  timestamppb.New(response.Tokens.AccessExpiresAt),
			RefreshExpiresAt: timestamppb.New(response.Tokens.RefreshExpiresAt),
			IsRevoked:        response.Tokens.IsRevoked,
			CreatedAt:        timestamppb.New(response.Tokens.CreatedAt),
		},
	}, nil
}

// SignOut handles user logout
func (h *AuthHandler) SignOut(ctx context.Context, req *proto.SignOutRequest) (*proto.Empty, error) {
	h.logger.Info(ctx, "Processing SignOut request")

	// Call domain service
	err := h.authService.SignOut(req.AccessToken)
	if err != nil {
		h.logger.Error(ctx, err, "SignOut failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "signout failed: %v", err)
	}

	return &proto.Empty{}, nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.TokenResponse, error) {
	h.logger.Info(ctx, "Processing RefreshToken request")

	// Call domain service
	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.logger.Error(ctx, err, "RefreshToken failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "token refresh failed: %v", err)
	}

	// Convert domain response to proto
	return &proto.TokenResponse{
		Tokens: &proto.UserToken{
			Id:               response.Tokens.ID.String(),
			UserId:           response.Tokens.UserID.String(),
			AccessToken:      response.Tokens.AccessToken,
			RefreshToken:     response.Tokens.RefreshToken,
			AccessExpiresAt:  timestamppb.New(response.Tokens.AccessExpiresAt),
			RefreshExpiresAt: timestamppb.New(response.Tokens.RefreshExpiresAt),
			IsRevoked:        response.Tokens.IsRevoked,
			CreatedAt:        timestamppb.New(response.Tokens.CreatedAt),
		},
	}, nil
}

// RevokeToken handles token revocation
func (h *AuthHandler) RevokeToken(ctx context.Context, req *proto.RevokeTokenRequest) (*proto.Empty, error) {
	h.logger.Info(ctx, "Processing RevokeToken request")

	// Call domain service
	err := h.authService.RevokeToken(req.AccessToken)
	if err != nil {
		h.logger.Error(ctx, err, "RevokeToken failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "token revocation failed: %v", err)
	}

	return &proto.Empty{}, nil
}

// ValidateToken handles token validation
func (h *AuthHandler) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	h.logger.Info(ctx, "Processing ValidateToken request")

	// Call domain service
	user, err := h.authService.ValidateToken(ctx, req.Token, h.jwtAccessTokenSecret)
	if err != nil {
		h.logger.Error(ctx, err, "ValidateToken failed", 401)
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

// ListUsers handles user listing
func (h *AuthHandler) ListUsers(ctx context.Context, req *proto.ListUsersRequest) (*proto.ListUsersResponse, error) {
	h.logger.Info(ctx, "Processing ListUsers request", map[string]any{
		"page":  req.Page,
		"limit": req.Limit,
	})

	// Call domain service
	users, total, err := h.userService.ListUsers(int(req.Page), int(req.Limit))
	if err != nil {
		h.logger.Error(ctx, err, "ListUsers failed", 500)
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}

	// Convert domain response to proto
	protoUsers := make([]*proto.User, len(users))
	for i, user := range users {
		protoUsers[i] = &proto.User{
			Id:        user.ID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: timestamppb.New(user.CreatedAt),
			UpdatedAt: timestamppb.New(user.UpdatedAt),
		}
	}

	return &proto.ListUsersResponse{
		Users: protoUsers,
		Total: int32(total),
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}
