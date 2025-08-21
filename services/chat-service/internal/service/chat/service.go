package chat

import (
	"context"
	"fmt"

	"chat-service/configs"
	"chat-service/internal/domain"
	"chat-service/internal/services/openai"
	"chat-service/storage"
	zlog "packages/logger"
)

// Service represents the chat service
type Service interface {
	SendMessage(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error)
	GetHistory(ctx context.Context, req *domain.GetHistoryRequest) (*domain.GetHistoryResponse, error)
	ListConversations(ctx context.Context, req *domain.ListConversationsRequest) (*domain.ListConversationsResponse, error)
	CreateConversation(ctx context.Context, userID, title string) (*domain.Conversation, error)
	ChatWithAI(ctx context.Context, userID, message, conversationID, model string, temperature float64, maxTokens int) (*domain.ChatResponse, error)
}

// service implements the chat service
type service struct {
	openaiClient openai.Client
	logger       *zlog.Logger
	config       *configs.Config
	storage      storage.Repository
}

// NewService creates a new chat service
func NewService(openaiClient openai.Client, logger *zlog.Logger, config *configs.Config, storage storage.Repository) Service {
	return &service{
		openaiClient: openaiClient,
		logger:       logger,
		config:       config,
		storage:      storage,
	}
}

// SendMessage sends a message and stores it
func (s *service) SendMessage(ctx context.Context, req *domain.ChatRequest) (*domain.ChatResponse, error) {
	s.logger.Info(ctx, "Sending message", map[string]interface{}{
		"user_id":         req.UserID,
		"conversation_id": req.ConversationID,
		"message_length":  len(req.Message),
	})

	// Create or get conversation ID FIRST
	conversationID := req.ConversationID
	if conversationID == "" {
		conversation := domain.NewConversation(req.UserID, "New Conversation")
		// Store the conversation BEFORE creating the message
		_, err := s.storage.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, fmt.Errorf("failed to store conversation: %w", err)
		}
		conversationID = conversation.ID
	} else {
		// Validate that the provided conversation exists and belongs to the user
		conversation, err := s.storage.GetConversationByID(ctx, conversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get conversation: %w", err)
		}
		if conversation == nil {
			return nil, fmt.Errorf("conversation not found: %s", conversationID)
		}
		if conversation.UserID != req.UserID {
			return nil, fmt.Errorf("conversation does not belong to user: %s", conversationID)
		}
	}

	// Create a new message with the conversation ID
	message := domain.NewMessage(req.UserID, conversationID, req.Message, "user")

	// Store the message in the database
	_, err := s.storage.CreateMessage(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to store message: %w", err)
	}

	response := &domain.ChatResponse{
		Message:        message,
		ConversationID: conversationID,
		IsAIResponse:   false,
	}

	s.logger.Info(ctx, "Message sent successfully", map[string]interface{}{
		"message_id":      message.ID,
		"conversation_id": conversationID,
	})

	return response, nil
}

// GetHistory retrieves chat history for a conversation
func (s *service) GetHistory(ctx context.Context, req *domain.GetHistoryRequest) (*domain.GetHistoryResponse, error) {
	s.logger.Info(ctx, "Getting chat history", map[string]interface{}{
		"user_id":         req.UserID,
		"conversation_id": req.ConversationID,
		"limit":           req.Limit,
		"offset":          req.Offset,
	})

	// Retrieve messages from the database
	messages, err := s.storage.GetMessagesByConversationID(ctx, req.ConversationID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Get total count
	total, err := s.storage.CountMessagesByConversationID(ctx, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get message count: %w", err)
	}

	// Convert []domain.Message to []*domain.Message
	var messagePtrs []*domain.Message
	for i := range messages {
		messagePtrs = append(messagePtrs, &messages[i])
	}

	response := &domain.GetHistoryResponse{
		Messages:       messagePtrs,
		Total:          total,
		ConversationID: req.ConversationID,
	}

	s.logger.Info(ctx, "Chat history retrieved", map[string]interface{}{
		"conversation_id": req.ConversationID,
		"total_messages":  total,
	})

	return response, nil
}

// ListConversations lists user conversations
func (s *service) ListConversations(ctx context.Context, req *domain.ListConversationsRequest) (*domain.ListConversationsResponse, error) {
	s.logger.Info(ctx, "Listing conversations", map[string]interface{}{
		"user_id": req.UserID,
		"limit":   req.Limit,
		"offset":  req.Offset,
	})

	// Retrieve conversations from the database
	conversations, err := s.storage.GetConversationsByUserID(ctx, req.UserID, req.Limit, req.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	// Get total count
	total, err := s.storage.CountConversationsByUserID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation count: %w", err)
	}

	// Convert []domain.Conversation to []*domain.Conversation
	var conversationPtrs []*domain.Conversation
	for i := range conversations {
		conversationPtrs = append(conversationPtrs, &conversations[i])
	}

	response := &domain.ListConversationsResponse{
		Conversations: conversationPtrs,
		Total:         total,
	}

	s.logger.Info(ctx, "Conversations listed", map[string]interface{}{
		"user_id":             req.UserID,
		"total_conversations": total,
	})

	return response, nil
}

// CreateConversation creates a new conversation
func (s *service) CreateConversation(ctx context.Context, userID, title string) (*domain.Conversation, error) {
	s.logger.Info(ctx, "Creating conversation", map[string]interface{}{
		"user_id": userID,
		"title":   title,
	})

	conversation := domain.NewConversation(userID, title)

	// Store the conversation in the database
	_, err := s.storage.CreateConversation(ctx, conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to store conversation: %w", err)
	}

	s.logger.Info(ctx, "Conversation created successfully", map[string]interface{}{
		"conversation_id": conversation.ID,
		"user_id":         userID,
	})

	return conversation, nil
}

// ChatWithAI sends a message to OpenAI and returns the AI response
func (s *service) ChatWithAI(ctx context.Context, userID, message, conversationID, model string, temperature float64, maxTokens int) (*domain.ChatResponse, error) {
	s.logger.Info(ctx, "Chatting with AI", map[string]interface{}{
		"user_id":         userID,
		"conversation_id": conversationID,
		"model":           model,
		"temperature":     temperature,
		"max_tokens":      maxTokens,
	})

	// Create or get conversation ID
	if conversationID == "" {
		conversation := domain.NewConversation(userID, "AI Chat")
		conversationID = conversation.ID
		// Store the conversation
		_, err := s.storage.CreateConversation(ctx, conversation)
		if err != nil {
			return nil, fmt.Errorf("failed to store conversation: %w", err)
		}
	}

	// Store user message
	userMsg := domain.NewMessage(userID, conversationID, message, "user")
	_, err := s.storage.CreateMessage(ctx, userMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to store user message: %w", err)
	}

	// Prepare messages for OpenAI
	openaiMessages := []openai.Message{
		{
			Role:    "user",
			Content: message,
		},
	}

	// Call OpenAI API
	aiResponse, err := s.openaiClient.ChatCompletion(ctx, openaiMessages, model, temperature, maxTokens)
	if err != nil {
		s.logger.Error(ctx, err, "Failed to get AI response", 500)
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Get AI message content
	aiMessageContent := aiResponse.GetFirstChoiceContent()
	if aiMessageContent == "" {
		return nil, fmt.Errorf("no AI response content received")
	}

	// Store AI message
	aiMsg := domain.NewMessage(userID, conversationID, aiMessageContent, "assistant")
	_, err = s.storage.CreateMessage(ctx, aiMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to store AI message: %w", err)
	}

	response := &domain.ChatResponse{
		Message:        aiMsg,
		ConversationID: conversationID,
		IsAIResponse:   true,
	}

	s.logger.Info(ctx, "AI chat completed successfully", map[string]interface{}{
		"conversation_id": conversationID,
		"tokens_used":     aiResponse.GetTotalTokens(),
		"model_used":      aiResponse.Model,
	})

	return response, nil
}
