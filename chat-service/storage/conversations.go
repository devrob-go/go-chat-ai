package storage

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"chat-service/internal/domain"

	"github.com/google/uuid"
)

// Named queries
const (
	insertConversationQuery = `
		INSERT INTO conversations (
			id,
			user_id,
			title,
			created_at,
			updated_at
		) VALUES (
			:id,
			:user_id,
			:title,
			:created_at,
			:updated_at
		)
		RETURNING id, user_id, title, created_at, updated_at
	`

	getConversationByIDQuery = `
		SELECT 
			id,
			user_id,
			title,
			created_at,
			updated_at
		FROM conversations
		WHERE id = :id
	`

	getConversationsByUserIDQuery = `
		SELECT 
			id,
			user_id,
			title,
			created_at,
			updated_at
		FROM conversations
		WHERE user_id = :user_id
		ORDER BY updated_at DESC
		LIMIT :limit OFFSET :offset
	`

	countConversationsByUserIDQuery = `
		SELECT COUNT(*) FROM conversations WHERE user_id = :user_id
	`

	updateConversationTitleQuery = `
		UPDATE conversations 
		SET title = :title, updated_at = :updated_at
		WHERE id = :id AND user_id = :user_id
		RETURNING id, user_id, title, created_at, updated_at
	`

	deleteConversationQuery = `
		DELETE FROM conversations 
		WHERE id = :id AND user_id = :user_id
	`
)

// CreateConversation inserts a new conversation into the database
func (db *DB) CreateConversation(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error) {
	if conversation.ID == "" {
		conversation.ID = uuid.New().String()
	}

	stmt, err := db.PrepareNamedContext(ctx, insertConversationQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare insert failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	var newConversation domain.Conversation
	if err := stmt.GetContext(ctx, &newConversation, conversation); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "insert failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "conversation created successfully", map[string]any{
		"conversation_id": newConversation.ID,
		"user_id":         newConversation.UserID,
		"title":           newConversation.Title,
	})

	return &newConversation, nil
}

// GetConversationByID retrieves a conversation by ID
func (db *DB) GetConversationByID(ctx context.Context, id string) (*domain.Conversation, error) {
	params := map[string]any{
		"id": id,
	}

	var conversation domain.Conversation
	stmt, err := db.PrepareNamedContext(ctx, getConversationByIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &conversation, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "conversation not found", map[string]any{
				"conversation_id": id,
			})
			return nil, errors.New("conversation not found")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	return &conversation, nil
}

// GetConversationsByUserID retrieves conversations for a specific user with pagination
func (db *DB) GetConversationsByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Conversation, error) {
	params := map[string]any{
		"user_id": userID,
		"limit":   limit,
		"offset":  offset,
	}

	var conversations []domain.Conversation
	stmt, err := db.PrepareNamedContext(ctx, getConversationsByUserIDQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare select failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	if err := stmt.SelectContext(ctx, &conversations, params); err != nil {
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "select failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "conversations retrieved successfully", map[string]any{
		"user_id":            userID,
		"count":              len(conversations),
		"limit":              limit,
		"offset":             offset,
	})

	return conversations, nil
}

// CountConversationsByUserID returns the total number of conversations for a user
func (db *DB) CountConversationsByUserID(ctx context.Context, userID string) (int, error) {
	params := map[string]any{
		"user_id": userID,
	}

	var count int
	stmt, err := db.PrepareNamedContext(ctx, countConversationsByUserIDQuery)
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

// UpdateConversationTitle updates the title of a conversation
func (db *DB) UpdateConversationTitle(ctx context.Context, id, userID, title string) (*domain.Conversation, error) {
	params := map[string]any{
		"id":         id,
		"user_id":    userID,
		"title":      title,
		"updated_at": domain.NewConversation(userID, title).UpdatedAt, // This will be overwritten
	}

	stmt, err := db.PrepareNamedContext(ctx, updateConversationTitleQuery)
	if err != nil {
		db.logger.Error(ctx, err, "prepare update failed", http.StatusInternalServerError)
		return nil, err
	}
	defer stmt.Close()

	var conversation domain.Conversation
	if err := stmt.GetContext(ctx, &conversation, params); err != nil {
		if err == sql.ErrNoRows {
			db.logger.Info(ctx, "conversation not found or user not authorized", map[string]any{
				"conversation_id": id,
				"user_id":         userID,
			})
			return nil, errors.New("conversation not found or user not authorized")
		}
		status, mappedErr := HandlePgError(err)
		db.logger.Error(ctx, mappedErr, "update failed", status)
		return nil, mappedErr
	}

	db.logger.Info(ctx, "conversation title updated successfully", map[string]any{
		"conversation_id": id,
		"user_id":         userID,
		"new_title":       title,
	})

	return &conversation, nil
}

// DeleteConversation deletes a conversation (this will cascade delete messages)
func (db *DB) DeleteConversation(ctx context.Context, id, userID string) error {
	params := map[string]any{
		"id":      id,
		"user_id": userID,
	}

	stmt, err := db.PrepareNamedContext(ctx, deleteConversationQuery)
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
		db.logger.Info(ctx, "conversation not found or user not authorized", map[string]any{
			"conversation_id": id,
			"user_id":         userID,
		})
		return errors.New("conversation not found or user not authorized")
	}

	db.logger.Info(ctx, "conversation deleted successfully", map[string]any{
		"conversation_id": id,
		"user_id":         userID,
		"rows_affected":   rowsAffected,
	})

	return nil
}
