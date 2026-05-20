package rag

import (
	"context"
	"testing"
)

type fakeEmbedder struct{}

func (fakeEmbedder) Embed(context.Context, string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

type fakeStore struct {
	savedChunks []string
}

func (f *fakeStore) SaveChunkEmbedding(_ context.Context, chunkID string, _ []float32) error {
	f.savedChunks = append(f.savedChunks, chunkID)
	return nil
}

func (f *fakeStore) ListChunksWithoutEmbedding(context.Context, int) ([]RetrievedChunk, error) {
	return []RetrievedChunk{
		{ChunkID: "chunk-1", Content: "first"},
		{ChunkID: "chunk-2", Content: "second"},
	}, nil
}

func (f *fakeStore) SearchChunks(context.Context, string, []string, []float32, int) ([]RetrievedChunk, error) {
	return []RetrievedChunk{{ChunkID: "chunk-1", Content: "first", Similarity: 0.9}}, nil
}

func TestIndexMissingEmbeddings(t *testing.T) {
	store := &fakeStore{}
	service := NewService(store, fakeEmbedder{})

	indexed, err := service.IndexMissingEmbeddings(context.Background(), 10)
	if err != nil {
		t.Fatalf("index missing embeddings: %v", err)
	}

	if indexed != 2 {
		t.Fatalf("expected 2 indexed chunks, got %d", indexed)
	}
}

func TestRetrieveEmbedsQueryAndSearches(t *testing.T) {
	service := NewService(&fakeStore{}, fakeEmbedder{})

	chunks, err := service.Retrieve(context.Background(), "user-1", []string{"document-1"}, "query", 3)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}

	if len(chunks) != 1 || chunks[0].ChunkID != "chunk-1" {
		t.Fatalf("expected retrieved chunk, got %#v", chunks)
	}
}
