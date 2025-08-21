package storage

import (
	"context"

	"chat-service/internal/domain"
)

// Repository defines the interface for chat storage operations
type Repository interface {
	// Conversation operations
	CreateConversation(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error)
	GetConversationByID(ctx context.Context, id string) (*domain.Conversation, error)
	GetConversationsByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Conversation, error)
	CountConversationsByUserID(ctx context.Context, userID string) (int, error)
	UpdateConversationTitle(ctx context.Context, id, userID, title string) (*domain.Conversation, error)
	DeleteConversation(ctx context.Context, id, userID string) error

	// Message operations
	CreateMessage(ctx context.Context, message *domain.Message) (*domain.Message, error)
	GetMessageByID(ctx context.Context, id string) (*domain.Message, error)
	GetMessagesByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, error)
	CountMessagesByConversationID(ctx context.Context, conversationID string) (int, error)
	GetMessagesByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Message, error)
	CountMessagesByUserID(ctx context.Context, userID string) (int, error)
	UpdateMessageContent(ctx context.Context, id, userID, content string) (*domain.Message, error)
	DeleteMessage(ctx context.Context, id, userID string) error
}

// Ensure DB implements Repository interface
var _ Repository = (*DB)(nil)
