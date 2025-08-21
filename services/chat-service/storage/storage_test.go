package storage

import (
	"testing"
	"time"

	"chat-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestNewMessage(t *testing.T) {
	userID := "test-user-123"
	conversationID := "test-conversation-456"
	content := "Hello, world!"
	role := "user"

	message := domain.NewMessage(userID, conversationID, content, role)

	assert.NotEmpty(t, message.ID)
	assert.Equal(t, userID, message.UserID)
	assert.Equal(t, conversationID, message.ConversationID)
	assert.Equal(t, content, message.Content)
	assert.Equal(t, role, message.Role)
	assert.WithinDuration(t, time.Now(), message.CreatedAt, 2*time.Second)
	assert.WithinDuration(t, time.Now(), message.UpdatedAt, 2*time.Second)
}

func TestNewConversation(t *testing.T) {
	userID := "test-user-123"
	title := "Test Conversation"

	conversation := domain.NewConversation(userID, title)

	assert.NotEmpty(t, conversation.ID)
	assert.Equal(t, userID, conversation.UserID)
	assert.Equal(t, title, conversation.Title)
	assert.WithinDuration(t, time.Now(), conversation.CreatedAt, 2*time.Second)
	assert.WithinDuration(t, time.Now(), conversation.UpdatedAt, 2*time.Second)
}

func TestMessageValidation(t *testing.T) {
	tests := []struct {
		name    string
		message *domain.Message
		isValid bool
	}{
		{
			name: "valid message",
			message: &domain.Message{
				ID:             "msg-123",
				UserID:         "user-123",
				ConversationID: "conv-123",
				Content:        "Hello",
				Role:           "user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			isValid: true,
		},
		{
			name: "missing ID",
			message: &domain.Message{
				UserID:         "user-123",
				ConversationID: "conv-123",
				Content:        "Hello",
				Role:           "user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			isValid: false,
		},
		{
			name: "missing user ID",
			message: &domain.Message{
				ID:             "msg-123",
				ConversationID: "conv-123",
				Content:        "Hello",
				Role:           "user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			isValid: false,
		},
		{
			name: "missing conversation ID",
			message: &domain.Message{
				ID:        "msg-123",
				UserID:    "user-123",
				Content:   "Hello",
				Role:      "user",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: false,
		},
		{
			name: "missing content",
			message: &domain.Message{
				ID:             "msg-123",
				UserID:         "user-123",
				ConversationID: "conv-123",
				Role:           "user",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			isValid: false,
		},
		{
			name: "missing role",
			message: &domain.Message{
				ID:             "msg-123",
				UserID:         "user-123",
				ConversationID: "conv-123",
				Content:        "Hello",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.message.ID)
				assert.NotEmpty(t, tt.message.UserID)
				assert.NotEmpty(t, tt.message.ConversationID)
				assert.NotEmpty(t, tt.message.Content)
				assert.NotEmpty(t, tt.message.Role)
				assert.False(t, tt.message.CreatedAt.IsZero())
				assert.False(t, tt.message.UpdatedAt.IsZero())
			} else {
				// At least one required field should be missing
				hasMissingField := tt.message.ID == "" ||
					tt.message.UserID == "" ||
					tt.message.ConversationID == "" ||
					tt.message.Content == "" ||
					tt.message.Role == ""
				assert.True(t, hasMissingField, "Message should have missing required fields")
			}
		})
	}
}

func TestConversationValidation(t *testing.T) {
	tests := []struct {
		name         string
		conversation *domain.Conversation
		isValid      bool
	}{
		{
			name: "valid conversation",
			conversation: &domain.Conversation{
				ID:        "conv-123",
				UserID:    "user-123",
				Title:     "Test Conversation",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: true,
		},
		{
			name: "missing ID",
			conversation: &domain.Conversation{
				UserID:    "user-123",
				Title:     "Test Conversation",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: false,
		},
		{
			name: "missing user ID",
			conversation: &domain.Conversation{
				ID:        "conv-123",
				Title:     "Test Conversation",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: false,
		},
		{
			name: "missing title",
			conversation: &domain.Conversation{
				ID:        "conv-123",
				UserID:    "user-123",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				assert.NotEmpty(t, tt.conversation.ID)
				assert.NotEmpty(t, tt.conversation.UserID)
				assert.NotEmpty(t, tt.conversation.Title)
				assert.False(t, tt.conversation.CreatedAt.IsZero())
				assert.False(t, tt.conversation.UpdatedAt.IsZero())
			} else {
				// At least one required field should be missing
				hasMissingField := tt.conversation.ID == "" ||
					tt.conversation.UserID == "" ||
					tt.conversation.Title == ""
				assert.True(t, hasMissingField, "Conversation should have missing required fields")
			}
		})
	}
}

func TestRepositoryInterface(t *testing.T) {
	// This test ensures that the DB struct implements the Repository interface
	var _ Repository = (*DB)(nil)
}

func TestConfigDefaults(t *testing.T) {
	// Test that default values are set correctly
	assert.Equal(t, DefaultMigrationsDir, "./storage/migrations")
	assert.Equal(t, DefaultConnTimeout, 5*time.Second)
	assert.Equal(t, DefaultMaxOpenConns, 25)
	assert.Equal(t, DefaultMaxIdleConns, 10)
	assert.Equal(t, DefaultConnMaxLifetime, 30*time.Minute)
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are defined
	assert.NotNil(t, ErrUniqueViolation)
	assert.NotNil(t, ErrForeignKeyViolation)
	assert.NotNil(t, ErrNotNullViolation)
	assert.NotNil(t, ErrCheckViolation)
	assert.NotNil(t, ErrExclusionViolation)
}
