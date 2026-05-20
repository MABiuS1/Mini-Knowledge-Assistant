package rag

import (
	"context"
	"errors"
	"testing"
)

type fakeEmbedder struct {
	err error
}

func (f fakeEmbedder) Embed(context.Context, string) ([]float32, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

type fakeStore struct {
	savedChunks []string
	searchLimit int
	err         error
}

func (f *fakeStore) SaveChunkEmbedding(_ context.Context, chunkID string, _ []float32) error {
	if f.err != nil {
		return f.err
	}
	f.savedChunks = append(f.savedChunks, chunkID)
	return nil
}

func (f *fakeStore) ListChunksWithoutEmbedding(context.Context, int) ([]RetrievedChunk, error) {
	if f.err != nil {
		return nil, f.err
	}
	return []RetrievedChunk{
		{ChunkID: "chunk-1", Content: "first"},
		{ChunkID: "chunk-2", Content: "second"},
	}, nil
}

func (f *fakeStore) SearchChunks(_ context.Context, _ string, _ []string, _ []float32, limit int) ([]RetrievedChunk, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.searchLimit = limit
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
	store := &fakeStore{}
	service := NewService(store, fakeEmbedder{})

	chunks, err := service.Retrieve(context.Background(), "user-1", []string{"document-1"}, "query", 3)
	if err != nil {
		t.Fatalf("retrieve: %v", err)
	}

	if len(chunks) != 1 || chunks[0].ChunkID != "chunk-1" {
		t.Fatalf("expected retrieved chunk, got %#v", chunks)
	}

	if store.searchLimit != 3 {
		t.Fatalf("expected search limit 3, got %d", store.searchLimit)
	}
}

func TestIndexMissingEmbeddingsStopsOnEmbedError(t *testing.T) {
	embedErr := errors.New("embed failed")
	service := NewService(&fakeStore{}, fakeEmbedder{err: embedErr})

	indexed, err := service.IndexMissingEmbeddings(context.Background(), 10)
	if !errors.Is(err, embedErr) {
		t.Fatalf("expected embed error, got %v", err)
	}

	if indexed != 0 {
		t.Fatalf("expected 0 indexed chunks, got %d", indexed)
	}
}

func TestRetrieveReturnsEmbedError(t *testing.T) {
	embedErr := errors.New("embed failed")
	service := NewService(&fakeStore{}, fakeEmbedder{err: embedErr})

	_, err := service.Retrieve(context.Background(), "user-1", []string{"document-1"}, "query", 3)
	if !errors.Is(err, embedErr) {
		t.Fatalf("expected embed error, got %v", err)
	}
}
