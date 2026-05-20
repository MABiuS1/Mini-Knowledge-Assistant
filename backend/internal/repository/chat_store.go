package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mabius/knowledge-assistant/backend/internal/chat"
)

type ChatStore struct {
	db *sql.DB
}

func NewChatStore(db *sql.DB) *ChatStore {
	return &ChatStore{db: db}
}

func (s *ChatStore) CreateConversation(ctx context.Context, userID string, title string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO conversations (user_id, title)
		VALUES ($1::uuid, $2)
		RETURNING id::text
	`, userID, title).Scan(&id)
	return id, err
}

func (s *ChatStore) VerifyConversationOwner(ctx context.Context, userID string, conversationID string) error {
	var exists bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM conversations
			WHERE id = $1::uuid
			  AND user_id = $2::uuid
		)
	`, conversationID, userID).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		return chat.ErrConversationNotFound
	}

	return nil
}

func (s *ChatStore) ListMessages(ctx context.Context, conversationID string, limit int) ([]chat.Message, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, role, content, prompt_tokens, completion_tokens, total_tokens
		FROM (
			SELECT id, role, content, prompt_tokens, completion_tokens, total_tokens, created_at
			FROM messages
			WHERE conversation_id = $1::uuid
			ORDER BY created_at DESC
			LIMIT $2
		) recent_messages
		ORDER BY created_at ASC
	`, conversationID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := []chat.Message{}
	for rows.Next() {
		var message chat.Message
		if err := rows.Scan(
			&message.ID,
			&message.Role,
			&message.Content,
			&message.Usage.PromptTokens,
			&message.Usage.CompletionTokens,
			&message.Usage.TotalTokens,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (s *ChatStore) CreateMessage(ctx context.Context, conversationID string, role chat.Role, content string, usage chat.Usage) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var id string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO messages (conversation_id, role, content, prompt_tokens, completion_tokens, total_tokens)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id::text
	`, conversationID, string(role), content, usage.PromptTokens, usage.CompletionTokens, usage.TotalTokens).Scan(&id)
	if err != nil {
		return "", err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE conversations
		SET updated_at = now()
		WHERE id = $1::uuid
	`, conversationID); err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return id, nil
}

func (s *ChatStore) CreateUsageEvent(ctx context.Context, params chat.UsageEventParams) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO usage_events (
			user_id,
			conversation_id,
			message_id,
			provider,
			model,
			prompt_tokens,
			completion_tokens,
			total_tokens
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8)
	`,
		params.UserID,
		params.ConversationID,
		params.MessageID,
		params.Provider,
		params.Model,
		params.Usage.PromptTokens,
		params.Usage.CompletionTokens,
		params.Usage.TotalTokens,
	)
	return err
}

func (s *ChatStore) SumConversationUsage(ctx context.Context, conversationID string) (chat.Usage, error) {
	var usage chat.Usage
	err := s.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(prompt_tokens), 0),
			COALESCE(SUM(completion_tokens), 0),
			COALESCE(SUM(total_tokens), 0)
		FROM usage_events
		WHERE conversation_id = $1::uuid
	`, conversationID).Scan(&usage.PromptTokens, &usage.CompletionTokens, &usage.TotalTokens)
	if errors.Is(err, sql.ErrNoRows) {
		return chat.Usage{}, nil
	}

	return usage, err
}
