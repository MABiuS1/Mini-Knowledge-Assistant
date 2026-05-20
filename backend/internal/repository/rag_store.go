package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"github.com/mabius/knowledge-assistant/backend/internal/rag"
)

type RAGStore struct {
	db *sql.DB
}

func NewRAGStore(db *sql.DB) *RAGStore {
	return &RAGStore{db: db}
}

func (s *RAGStore) SaveChunkEmbedding(ctx context.Context, chunkID string, embedding []float32) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE document_chunks
		SET embedding = $2::vector
		WHERE id = $1::uuid
	`, chunkID, vectorLiteral(embedding))
	return err
}

func (s *RAGStore) ListChunksWithoutEmbedding(ctx context.Context, limit int) ([]rag.RetrievedChunk, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			document_chunks.id::text,
			document_chunks.document_id::text,
			documents.original_name,
			document_chunks.chunk_index,
			document_chunks.content
		FROM document_chunks
		INNER JOIN documents ON documents.id = document_chunks.document_id
		WHERE document_chunks.embedding IS NULL
		ORDER BY document_chunks.created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := []rag.RetrievedChunk{}
	for rows.Next() {
		var chunk rag.RetrievedChunk
		if err := rows.Scan(&chunk.ChunkID, &chunk.DocumentID, &chunk.FileName, &chunk.ChunkIndex, &chunk.Content); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}

	return chunks, rows.Err()
}

func (s *RAGStore) SearchChunks(ctx context.Context, userID string, documentIDs []string, embedding []float32, limit int) ([]rag.RetrievedChunk, error) {
	if limit <= 0 {
		limit = 5
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT
			document_chunks.id::text,
			document_chunks.document_id::text,
			documents.original_name,
			document_chunks.chunk_index,
			document_chunks.content,
			1 - (document_chunks.embedding <=> $3::vector) AS similarity
		FROM document_chunks
		INNER JOIN documents ON documents.id = document_chunks.document_id
		WHERE documents.user_id = $1::uuid
		  AND document_chunks.embedding IS NOT NULL
		  AND (cardinality($2::uuid[]) = 0 OR documents.id = ANY($2::uuid[]))
		ORDER BY document_chunks.embedding <=> $3::vector
		LIMIT $4
	`, userID, documentIDs, vectorLiteral(embedding), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chunks := []rag.RetrievedChunk{}
	for rows.Next() {
		var chunk rag.RetrievedChunk
		if err := rows.Scan(
			&chunk.ChunkID,
			&chunk.DocumentID,
			&chunk.FileName,
			&chunk.ChunkIndex,
			&chunk.Content,
			&chunk.Similarity,
		); err != nil {
			return nil, err
		}
		chunks = append(chunks, chunk)
	}

	return chunks, rows.Err()
}

func vectorLiteral(values []float32) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatFloat(float64(value), 'f', -1, 32))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
