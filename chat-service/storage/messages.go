package storage

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"chat-service/internal/domain"

	"github.com/google/uuid"
)

// Named queries
const (
	insertMessageQuery = `
		INSERT INTO messages (
			id,
			user_id,
			conversation_id,
			content,
			role,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:conversation_id,
			:content,
			:role,
			:created_at,
			:updated_at
		)
		RETURNING id, user_id, conversation_id, content, role, created_at, updated_at
	`

	getMessageByIDQuery = `
		SELECT 
			id,
			user_id,
			conversation_id,
			content,
			role,
			created_at,
			updated_at
		FROM messages
		WHERE id = :id
	`

	getMessagesByConversationIDQuery = `
		SELECT 
			id,
			user_id,
			conversation_id,
			content,
			role,
			created_at,
			updated_at
		FROM messages
		WHERE conversation_id = :conversation_id
		ORDER BY created_at ASC
		LIMIT :limit OFFSET :offset
	`

	countMessagesByConversationIDQuery = `
		SELECT COUNT(*) FROM messages WHERE conversation_id = :conversation_id
	`

	getMessagesByUserIDQuery = `
		SELECT 
			id,
			user_id,
			conversation_id,
			content,
			role,
			created_at,
			updated_at
		FROM messages
		WHERE user_id = :user_id
		ORDER BY created_at DESC
		LIMIT :limit OFFSET :offset
	`

	countMessagesByUserIDQuery = `
		SELECT COUNT(*) FROM messages WHERE user_id = :user_id
	`

	updateMessageContentQuery = `
		UPDATE messages 
		SET content = :content, updated_at = :updated_at
		WHERE id = :id AND user_id = :user_id
		RETURNING id, user_id, conversation_id, content, role, created_at, updated_at
	`

	deleteMessageQuery = `
		DELETE FROM messages 
		WHERE id = :id AND user_id = :user_id
	`
)

// CreateMessage inserts a new message into the database
func (db *DB) CreateMessage(ctx context.Context, message *domain.Message) (*domain.Message, error) {
	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	stmt, err := db.PrepareNamedContext(ctx, insertMessageQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare insert failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	var newMessage domain.Message
	if err := stmt.GetContext(ctx, &newMessage, message); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "insert failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "message created successfully", map[string]any{
		"message_id":      newMessage.ID,
		"user_id":         newMessage.UserID,
		"conversation_id": newMessage.ConversationID,
		"role":            newMessage.Role,
		"content_length":  len(newMessage.Content),
	})

	return &newMessage, nil
}

// GetMessageByID retrieves a message by ID
func (db *DB) GetMessageByID(ctx context.Context, id string) (*domain.Message, error) {
	params := map[string]any{
		"id": id,
	}

	var message domain.Message
	stmt, err := db.PrepareNamedContext(ctx, getMessageByIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &message, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "message not found", map[string]any{
				"message_id": id,
			})
			return nil, errors.New("message not found")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	return &message, nil
}

// GetMessagesByConversationID retrieves messages for a specific conversation with pagination
func (db *DB) GetMessagesByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]domain.Message, error) {
	params := map[string]any{
		"conversation_id": conversationID,
		"limit":           limit,
		"offset":          offset,
	}

	var messages []domain.Message
	stmt, err := db.PrepareNamedContext(ctx, getMessagesByConversationIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &messages, params); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "messages retrieved successfully", map[string]any{
		"conversation_id": conversationID,
		"count":           len(messages),
		"limit":           limit,
		"offset":          offset,
	})

	return messages, nil
}

// CountMessagesByConversationID returns the total number of messages in a conversation
func (db *DB) CountMessagesByConversationID(ctx context.Context, conversationID string) (int, error) {
	params := map[string]any{
		"conversation_id": conversationID,
	}

	var count int
	stmt, err := db.PrepareNamedContext(ctx, countMessagesByConversationIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare count failed", http.StatusInternalServerError)
		return 0, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &count, params); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "count failed", status)
		return 0, mappedErr
	}

	return count, nil
}

// GetMessagesByUserID retrieves messages for a specific user with pagination
func (db *DB) GetMessagesByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Message, error) {
	params := map[string]any{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	var messages []domain.Message
	stmt, err := db.PrepareNamedContext(ctx, getMessagesByUserIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &messages, params); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "user messages retrieved successfully", map[string]any{
		"user_id": userID,
		"count":   len(messages),
		"limit":   limit,
		"offset":  offset,
	})

	return messages, nil
}

// CountMessagesByUserID returns the total number of messages for a user
func (db *DB) CountMessagesByUserID(ctx context.Context, userID string) (int, error) {
	params := map[string]any{
		"user_id": userID,
	}

	var count int
	stmt, err := db.PrepareNamedContext(ctx, countMessagesByUserIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare count failed", http.StatusInternalServerError)
		return 0, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &count, params); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "count failed", status)
		return 0, mappedErr
	}

	return count, nil
}

// UpdateMessageContent updates the content of a message
func (db *DB) UpdateMessageContent(ctx context.Context, id, userID, content string) (*domain.Message, error) {
	params := map[string]any{
		"id":         id,
		"user_id":    userID,
		"content":    content,
		"updated_at": time.Now(),
	}

	stmt, err := db.PrepareNamedContext(ctx, updateMessageContentQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare update failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	var message domain.Message
	if err := stmt.GetContext(ctx, &message, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "message not found or user not authorized", map[string]any{
				"message_id": id,
				"user_id":    userID,
			})
			return nil, errors.New("message not found or user not authorized")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "update failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "message content updated successfully", map[string]any{
		"message_id":         id,
		"user_id":            userID,
		"new_content_length": len(content),
	})

	return &message, nil
}

// DeleteMessage deletes a message
func (db *DB) DeleteMessage(ctx context.Context, id, userID string) error {
	params := map[string]any{
		"id":      id,
		"user_id": userID,
	}

	stmt, err := db.PrepareNamedContext(ctx, deleteMessageQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare delete failed", http.StatusInternalServerError)
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, params)
	if err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "delete failed", status)
		return mappedErr
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		db.logger.Error(ctx, err, "failed to get rows affected", http.StatusInternalServerError)
		return err
	}

	if rowsAffected == 0 {
		db.logger.Info(ctx, "message not found or user not authorized", map[string]any{
			"message_id": id,
			"user_id":    userID,
		})
		return errors.New("message not found or user not authorized")
	}

	db.logger.Info(ctx, "message deleted successfully", map[string]any{
		"message_id":    id,
		"user_id":       userID,
		"rows_affected": rowsAffected,
	})

	return nil
}
