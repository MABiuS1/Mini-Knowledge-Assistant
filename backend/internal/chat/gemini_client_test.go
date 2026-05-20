package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGeminiClientCompleteReturnsContentAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("unexpected api key %q", r.URL.Query().Get("key"))
		}

		if !strings.Contains(r.URL.Path, "/v1beta/models/gemini-test:generateContent") {
			t.Fatalf("unexpected path %q", r.URL.Path)
		}

		var request geminiGenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if request.SystemInstruction == nil {
			t.Fatal("expected system instruction")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"modelVersion": "gemini-test",
			"candidates": [{
				"content": {
					"parts": [{"text": "answer"}]
				}
			}],
			"usageMetadata": {
				"promptTokenCount": 3,
				"candidatesTokenCount": 4,
				"totalTokenCount": 7
			}
		}`))
	}))
	defer server.Close()

	client := NewGeminiClient(server.URL, "test-key", "gemini-test", time.Second)
	completion, err := client.Complete(context.Background(), []Message{
		{Role: RoleSystem, Content: "system"},
		{Role: RoleUser, Content: "hello"},
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	if completion.Content != "answer" {
		t.Fatalf("expected answer, got %q", completion.Content)
	}

	if completion.Usage.TotalTokens != 7 {
		t.Fatalf("expected total tokens, got %d", completion.Usage.TotalTokens)
	}
}

func TestGeminiClientCompleteReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key","status":"PERMISSION_DENIED"}}`))
	}))
	defer server.Close()

	client := NewGeminiClient(server.URL, "bad-key", "gemini-test", time.Second)
	_, err := client.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hello"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
