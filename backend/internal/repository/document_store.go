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

func (s *DocumentStore) CreateDocument(ctx context.Context, params document.CreateDocumentParams) (document.Document, error) {
	var doc document.Document
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO documents (user_id, original_name, stored_name, mime_type, size_bytes, status)
		VALUES ($1::uuid, $2, $3, $4, $5, 'uploaded')
		RETURNING id::text, original_name, stored_name, mime_type, size_bytes, status, chunk_count, created_at
	`, params.UserID, params.OriginalName, params.StoredName, params.MimeType, params.SizeBytes).
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

	if doc.CreatedAt.IsZero() {
		doc.CreatedAt = time.Now().UTC()
	}

	return doc, nil
}
