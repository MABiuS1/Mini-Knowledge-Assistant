package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGeminiEmbedderReturnsEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("unexpected api key %q", r.URL.Query().Get("key"))
		}

		if !strings.Contains(r.URL.Path, "/v1beta/models/gemini-embedding-001:embedContent") {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var request geminiEmbedRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if request.OutputDimensionality != 1536 {
			t.Fatalf("expected 1536 dimensions, got %d", request.OutputDimensionality)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"embedding":{"values":[0.1,0.2,0.3]}}`))
	}))
	defer server.Close()

	embedder := NewGeminiEmbedder(server.URL, "test-key", "gemini-embedding-001", 1536, time.Second)
	embedding, err := embedder.Embed(context.Background(), "hello")
	if err != nil {
		t.Fatalf("embed: %v", err)
	}

	if len(embedding) != 3 {
		t.Fatalf("expected embedding length 3, got %d", len(embedding))
	}
}
