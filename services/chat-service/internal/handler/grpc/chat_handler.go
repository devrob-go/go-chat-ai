package grpc

import (
	"context"
	"fmt"

	"chat-service/internal/domain"
	"chat-service/internal/services/chat"
	"chat-service/proto"
	zlog "packages/logger"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ChatHandler handles gRPC chat requests
type ChatHandler struct {
	proto.UnimplementedChatServiceServer
	chatService chat.Service
	logger      *zlog.Logger
}

// NewChatHandler creates a new chat handler
func NewChatHandler(chatService chat.Service, logger *zlog.Logger) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		logger:      logger,
	}
}

// SendMessage handles sending a message
func (h *ChatHandler) SendMessage(ctx context.Context, req *proto.ChatRequest) (*proto.ChatResponse, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error(ctx, fmt.Errorf("user_id not found in context"), "Failed to extract user_id from context", 500)
		return nil, status.Errorf(codes.Internal, "authentication error: user_id not found in context")
	}

	// Validate user ID is not empty
	if userID == "" {
		h.logger.Error(ctx, fmt.Errorf("user_id is empty"), "User ID is empty", 400)
		return nil, status.Errorf(codes.InvalidArgument, "user_id cannot be empty")
	}

	h.logger.Info(ctx, "Handling SendMessage request", map[string]any{
		"user_id":         userID,
		"conversation_id": req.ConversationId,
		"message_length":  len(req.Message),
	})

	// Convert proto request to domain request
	domainReq := &domain.ChatRequest{
		UserID:         userID,
		Message:        req.Message,
		ConversationID: req.ConversationId,
	}

	// Validate the domain request
	if err := domainReq.Validate(); err != nil {
		h.logger.Error(ctx, err, "Validation failed", 400)
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	// Call chat service
	response, err := h.chatService.SendMessage(ctx, domainReq)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to send message", 500)
		return nil, status.Errorf(codes.Internal, "failed to send message: %v", err)
	}

	// Convert domain response to proto response
	protoResponse := &proto.ChatResponse{
		Message:        h.convertMessageToProto(response.Message),
		ConversationId: response.ConversationID,
		IsAiResponse:   response.IsAIResponse,
	}

	h.logger.Info(ctx, "Message sent successfully", map[string]any{
		"message_id":      response.Message.ID,
		"conversation_id": response.ConversationID,
	})

	return protoResponse, nil
}

// StreamMessages handles streaming messages
func (h *ChatHandler) StreamMessages(req *proto.StreamMessageRequest, stream proto.ChatService_StreamMessagesServer) error {
	ctx := stream.Context()

	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error(ctx, fmt.Errorf("user_id not found in context"), "Failed to extract user_id from context", 500)
		return status.Errorf(codes.Internal, "authentication error")
	}

	h.logger.Info(ctx, "Handling StreamMessages request", map[string]any{
		"user_id":         userID,
		"conversation_id": req.ConversationId,
	})

	// For now, we'll just send a single message to demonstrate the streaming
	// In a real implementation, you would stream actual messages from the database or real-time updates

	message := &proto.Message{
		Id:        "stream-msg-1",
		UserId:    userID,
		Content:   "This is a streamed message",
		Role:      "assistant",
		CreatedAt: timestamppb.Now(),
		UpdatedAt: timestamppb.Now(),
	}

	response := &proto.StreamMessageResponse{
		Message: message,
		IsEnd:   false,
	}

	if err := stream.Send(response); err != nil {
		h.logger.Error(ctx, err, "Failed to send stream message", 500)
		return status.Errorf(codes.Internal, "failed to send stream message: %v", err)
	}

	// Send end message
	endResponse := &proto.StreamMessageResponse{
		Message: nil,
		IsEnd:   true,
	}

	if err := stream.Send(endResponse); err != nil {
		h.logger.Error(ctx, err, "Failed to send end message", 500)
		return status.Errorf(codes.Internal, "failed to send end message: %v", err)
	}

	h.logger.Info(ctx, "Stream messages completed", map[string]any{
		"conversation_id": req.ConversationId,
	})

	return nil
}

// GetHistory handles getting chat history
func (h *ChatHandler) GetHistory(ctx context.Context, req *proto.GetHistoryRequest) (*proto.GetHistoryResponse, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error(ctx, fmt.Errorf("user_id not found in context"), "Failed to extract user_id from context", 500)
		return nil, status.Errorf(codes.Internal, "authentication error")
	}

	h.logger.Info(ctx, "Handling GetHistory request", map[string]any{
		"user_id":         userID,
		"conversation_id": req.ConversationId,
		"limit":           req.Limit,
		"offset":          req.Offset,
	})

	// Convert proto request to domain request
	domainReq := &domain.GetHistoryRequest{
		UserID:         userID,
		ConversationID: req.ConversationId,
		Limit:          int(req.Limit),
		Offset:         int(req.Offset),
	}

	// Call chat service
	response, err := h.chatService.GetHistory(ctx, domainReq)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to get chat history", 500)
		return nil, status.Errorf(codes.Internal, "failed to get chat history: %v", err)
	}

	// Convert domain response to proto response
	protoMessages := make([]*proto.Message, len(response.Messages))
	for i, msg := range response.Messages {
		protoMessages[i] = h.convertMessageToProto(msg)
	}

	protoResponse := &proto.GetHistoryResponse{
		Messages:       protoMessages,
		Total:          int32(response.Total),
		ConversationId: response.ConversationID,
	}

	h.logger.Info(ctx, "Chat history retrieved", map[string]any{
		"conversation_id": response.ConversationID,
		"total_messages":  response.Total,
	})

	return protoResponse, nil
}

// ChatWithAI handles chatting with AI
func (h *ChatHandler) ChatWithAI(ctx context.Context, req *proto.ChatWithAIRequest) (*proto.ChatWithAIResponse, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error(ctx, fmt.Errorf("user_id not found in context"), "Failed to extract user_id from context", 500)
		return nil, status.Errorf(codes.Internal, "authentication error")
	}

	h.logger.Info(ctx, "Handling ChatWithAI request", map[string]any{
		"user_id":         userID,
		"conversation_id": req.ConversationId,
		"model":           req.Model,
		"temperature":     req.Temperature,
		"max_tokens":      req.MaxTokens,
	})

	// Call chat service
	response, err := h.chatService.ChatWithAI(
		ctx,
		userID,
		req.Message,
		req.ConversationId,
		req.Model,
		float64(req.Temperature),
		int(req.MaxTokens),
	)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to chat with AI", 500)
		return nil, status.Errorf(codes.Internal, "failed to chat with AI: %v", err)
	}

	// Convert domain response to proto response
	protoResponse := &proto.ChatWithAIResponse{
		AiMessage:      response.Message.Content,
		ConversationId: response.ConversationID,
		ModelUsed:      req.Model,
		TokensUsed:     int32(0), // This would come from OpenAI response in real implementation
		CreatedAt:      timestamppb.Now(),
	}

	h.logger.Info(ctx, "AI chat completed successfully", map[string]any{
		"conversation_id": response.ConversationID,
		"model_used":      req.Model,
	})

	return protoResponse, nil
}

// ListConversations handles listing conversations
func (h *ChatHandler) ListConversations(ctx context.Context, req *proto.ListConversationsRequest) (*proto.ListConversationsResponse, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error(ctx, fmt.Errorf("user_id not found in context"), "Failed to extract user_id from context", 500)
		return nil, status.Errorf(codes.Internal, "authentication error")
	}

	h.logger.Info(ctx, "Handling ListConversations request", map[string]any{
		"user_id": userID,
		"limit":   req.Limit,
		"offset":  req.Offset,
	})

	// Convert proto request to domain request
	domainReq := &domain.ListConversationsRequest{
		UserID: userID,
		Limit:  int(req.Limit),
		Offset: int(req.Offset),
	}

	// Call chat service
	response, err := h.chatService.ListConversations(ctx, domainReq)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to list conversations", 500)
		return nil, status.Errorf(codes.Internal, "failed to list conversations: %v", err)
	}

	// Convert domain response to proto response
	protoConversations := make([]*proto.Conversation, len(response.Conversations))
	for i, conv := range response.Conversations {
		protoConversations[i] = h.convertConversationToProto(conv)
	}

	protoResponse := &proto.ListConversationsResponse{
		Conversations: protoConversations,
		Total:         int32(response.Total),
	}

	h.logger.Info(ctx, "Conversations listed", map[string]any{
		"user_id":             userID,
		"total_conversations": response.Total,
	})

	return protoResponse, nil
}

// CreateConversation handles creating a new conversation
func (h *ChatHandler) CreateConversation(ctx context.Context, req *proto.Conversation) (*proto.Conversation, error) {
	// Extract user ID from context (set by auth interceptor)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		h.logger.Error(ctx, fmt.Errorf("user_id not found in context"), "Failed to extract user_id from context", 500)
		return nil, status.Errorf(codes.Internal, "authentication error")
	}

	h.logger.Info(ctx, "Handling CreateConversation request", map[string]any{
		"user_id": userID,
		"title":   req.Title,
	})

	// Call chat service
	conversation, err := h.chatService.CreateConversation(ctx, userID, req.Title)
	if err != nil {
		h.logger.Error(ctx, err, "Failed to create conversation", 500)
		return nil, status.Errorf(codes.Internal, "failed to create conversation: %v", err)
	}

	// Convert domain response to proto response
	protoResponse := h.convertConversationToProto(conversation)

	h.logger.Info(ctx, "Conversation created successfully", map[string]any{
		"conversation_id": conversation.ID,
		"user_id":         conversation.UserID,
	})

	return protoResponse, nil
}

// Helper functions to convert between domain and proto types
func (h *ChatHandler) convertMessageToProto(msg *domain.Message) *proto.Message {
	if msg == nil {
		return nil
	}

	return &proto.Message{
		Id:        msg.ID,
		UserId:    msg.UserID,
		Content:   msg.Content,
		Role:      msg.Role,
		CreatedAt: timestamppb.New(msg.CreatedAt),
		UpdatedAt: timestamppb.New(msg.UpdatedAt),
	}
}

func (h *ChatHandler) convertConversationToProto(conv *domain.Conversation) *proto.Conversation {
	if conv == nil {
		return nil
	}

	return &proto.Conversation{
		Id:        conv.ID,
		Title:     conv.Title,
		CreatedAt: timestamppb.New(conv.CreatedAt),
		UpdatedAt: timestamppb.New(conv.UpdatedAt),
	}
}
