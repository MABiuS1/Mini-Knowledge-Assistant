package chat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAIClientCompleteReturnsContentAndUsage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("unexpected authorization header %q", r.Header.Get("Authorization"))
		}

		var request openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if request.Model != "test-model" {
			t.Fatalf("expected model test-model, got %q", request.Model)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"model": "test-model",
			"choices": [{"message": {"role": "assistant", "content": "answer"}}],
			"usage": {"prompt_tokens": 3, "completion_tokens": 4, "total_tokens": 7}
		}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "test-key", "test-model", time.Second)
	completion, err := client.Complete(context.Background(), []Message{
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

func TestOpenAIClientCompleteReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer server.Close()

	client := NewOpenAIClient(server.URL, "bad-key", "test-model", time.Second)
	_, err := client.Complete(context.Background(), []Message{{Role: RoleUser, Content: "hello"}})
	if err == nil {
		t.Fatal("expected error")
	}
}
