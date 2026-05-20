package httpapi

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mabius/knowledge-assistant/backend/internal/rag"
)

type fakeRAGService struct {
	chunks []rag.RetrievedChunk
	err    error
}

func (f fakeRAGService) Retrieve(context.Context, string, []string, string, int) ([]rag.RetrievedChunk, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.chunks, nil
}

func TestRAGRetrieveRouteRequiresAuth(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		RAGService:  fakeRAGService{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rag/retrieve", bytes.NewBufferString(`{"query":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestRAGRetrieveRouteReturnsChunks(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		RAGService: fakeRAGService{
			chunks: []rag.RetrievedChunk{{ChunkID: "chunk-1", Content: "hello", Similarity: 0.9}},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rag/retrieve", bytes.NewBufferString(`{"query":"hello","limit":3}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer session-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRAGRetrieveRouteRejectsEmptyQuery(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		RAGService:  fakeRAGService{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/rag/retrieve", bytes.NewBufferString(`{"query":" "}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer session-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}
