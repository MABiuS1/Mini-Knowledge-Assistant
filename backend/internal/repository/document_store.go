package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/mabius/knowledge-assistant/backend/internal/document"
)

type DocumentStore struct {
	db *sql.DB
}

func NewDocumentStore(db *sql.DB) *DocumentStore {
	return &DocumentStore{db: db}
}

func (s *DocumentStore) CreateDocument(ctx context.Context, params document.CreateDocumentParams, chunks []document.Chunk) (document.Document, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return document.Document{}, err
	}
	defer tx.Rollback()

	var doc document.Document
	err = tx.QueryRowContext(ctx, `
		INSERT INTO documents (user_id, original_name, stored_name, mime_type, size_bytes, status, chunk_count)
		VALUES ($1::uuid, $2, $3, $4, $5, 'ready', $6)
		RETURNING id::text, original_name, stored_name, mime_type, size_bytes, status, chunk_count, created_at
	`, params.UserID, params.OriginalName, params.StoredName, params.MimeType, params.SizeBytes, len(chunks)).
		Scan(
			&doc.ID,
			&doc.OriginalName,
			&doc.StoredName,
			&doc.MimeType,
			&doc.SizeBytes,
			&doc.Status,
			&doc.ChunkCount,
			&doc.CreatedAt,
		)
	if err != nil {
		return document.Document{}, err
	}

	for _, chunk := range chunks {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO document_chunks (document_id, chunk_index, content, token_count)
			VALUES ($1::uuid, $2, $3, $4)
		`, doc.ID, chunk.ChunkIndex, chunk.Content, chunk.TokenCount)
		if err != nil {
			return document.Document{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return document.Document{}, err
	}

	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	return doc, nil
}

func (s *DocumentStore) ListDocuments(ctx context.Context, userID string) ([]document.Document, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id::text, original_name, stored_name, mime_type, size_bytes, status, chunk_count, created_at
		FROM documents
		WHERE user_id = $1::uuid
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	documents := []document.Document{}
	for rows.Next() {
		var doc document.Document
		if err := rows.Scan(
			&doc.ID,
			&doc.OriginalName,
			&doc.StoredName,
			&doc.MimeType,
			&doc.SizeBytes,
			&doc.Status,
			&doc.ChunkCount,
			&doc.CreatedAt,
		); err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	return documents, rows.Err()
}
