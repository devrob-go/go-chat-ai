package domain

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// UUID validation regex pattern
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidateUUID checks if a string is a valid UUID
func ValidateUUID(id string) error {
	if id == "" {
		return fmt.Errorf("UUID cannot be empty")
	}
	if !uuidRegex.MatchString(id) {
		return fmt.Errorf("invalid UUID format: %s", id)
	}
	return nil
}

// Message represents a chat message
type Message struct {
	ID             string    `json:"id" db:"id"`
	UserID         string    `json:"user_id" db:"user_id"`
	ConversationID string    `json:"conversation_id" db:"conversation_id"`
	Content        string    `json:"content" db:"content"`
	Role           string    `json:"role" db:"role"` // "user", "assistant", "system"
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Conversation represents a chat conversation
type Conversation struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ChatRequest represents a request to send a message
type ChatRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	Message        string `json:"message" validate:"required,min=1,max=4000"`
	ConversationID string `json:"conversation_id,omitempty"`
}

// Validate validates the ChatRequest
func (r *ChatRequest) Validate() error {
	if err := ValidateUUID(r.UserID); err != nil {
		return fmt.Errorf("user_id: %w", err)
	}
	if r.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	if len(r.Message) > 4000 {
		return fmt.Errorf("message too long (max 4000 characters)")
	}
	if r.ConversationID != "" {
		if err := ValidateUUID(r.ConversationID); err != nil {
			return fmt.Errorf("conversation_id: %w", err)
		}
	}
	return nil
}

// ChatResponse represents a response from the chat
type ChatResponse struct {
	Message        *Message `json:"message"`
	ConversationID string   `json:"conversation_id"`
	IsAIResponse   bool     `json:"is_ai_response"`
}

// GetHistoryRequest represents a request to get chat history
type GetHistoryRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	ConversationID string `json:"conversation_id" validate:"required"`
	Limit          int    `json:"limit" validate:"min=1,max=100"`
	Offset         int    `json:"offset" validate:"min=0"`
}

// Validate validates the GetHistoryRequest
func (r *GetHistoryRequest) Validate() error {
	if err := ValidateUUID(r.UserID); err != nil {
		return fmt.Errorf("user_id: %w", err)
	}
	if err := ValidateUUID(r.ConversationID); err != nil {
		return fmt.Errorf("conversation_id: %w", err)
	}
	if r.Limit < 1 || r.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}
	if r.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}

// ListConversationsRequest represents a request to list conversations
type ListConversationsRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Limit  int    `json:"limit" validate:"min=1,max=100"`
	Offset int    `json:"offset" validate:"min=0"`
}

// Validate validates the ListConversationsRequest
func (r *ListConversationsRequest) Validate() error {
	if err := ValidateUUID(r.UserID); err != nil {
		return fmt.Errorf("user_id: %w", err)
	}
	if r.Limit < 1 || r.Limit > 100 {
		return fmt.Errorf("limit must be between 1 and 100")
	}
	if r.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}
	return nil
}

// GetHistoryResponse represents a response with chat history
type GetHistoryResponse struct {
	Messages       []*Message `json:"messages"`
	Total          int        `json:"total"`
	ConversationID string     `json:"conversation_id"`
}

// ListConversationsResponse represents a response with conversations
type ListConversationsResponse struct {
	Conversations []*Conversation `json:"conversations"`
	Total         int             `json:"total"`
}

// NewMessage creates a new message
func NewMessage(userID, conversationID, content, role string) *Message {
	now := time.Now()
	return &Message{
		ID:             uuid.New().String(),
		UserID:         userID,
		ConversationID: conversationID,
		Content:        content,
		Role:           role,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewConversation creates a new conversation
func NewConversation(userID, title string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:        uuid.New().String(),
		UserID:    userID,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Standard error response structure
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(errorType, message, code string) *ErrorResponse {
	return &ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    code,
	}
}

// NewErrorResponseWithDetails creates a new error response with additional details
func NewErrorResponseWithDetails(errorType, message, code string, details map[string]string) *ErrorResponse {
	return &ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    code,
		Details: details,
	}
}
