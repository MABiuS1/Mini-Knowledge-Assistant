package rag

import "context"

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

type RetrievedChunk struct {
	DocumentID string  `json:"documentId"`
	FileName   string  `json:"fileName"`
	ChunkID    string  `json:"chunkId"`
	ChunkIndex int     `json:"chunkIndex"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type Store interface {
	SaveChunkEmbedding(ctx context.Context, chunkID string, embedding []float32) error
	ListChunksWithoutEmbedding(ctx context.Context, limit int) ([]RetrievedChunk, error)
	SearchChunks(ctx context.Context, userID string, documentIDs []string, embedding []float32, limit int) ([]RetrievedChunk, error)
}

type Service struct {
	store    Store
	embedder Embedder
}

func NewService(store Store, embedder Embedder) *Service {
	return &Service{
		store:    store,
		embedder: embedder,
	}
}

func (s *Service) IndexMissingEmbeddings(ctx context.Context, limit int) (int, error) {
	chunks, err := s.store.ListChunksWithoutEmbedding(ctx, limit)
	if err != nil {
		return 0, err
	}

	indexed := 0
	for _, chunk := range chunks {
		embedding, err := s.embedder.Embed(ctx, chunk.Content)
		if err != nil {
			return indexed, err
		}

		if err := s.store.SaveChunkEmbedding(ctx, chunk.ChunkID, embedding); err != nil {
			return indexed, err
		}

		indexed++
	}

	return indexed, nil
}

func (s *Service) Retrieve(ctx context.Context, userID string, documentIDs []string, query string, limit int) ([]RetrievedChunk, error) {
	embedding, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	return s.store.SearchChunks(ctx, userID, documentIDs, embedding, limit)
}
