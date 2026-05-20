package httpapi

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mabius/knowledge-assistant/backend/internal/chat"
)

type fakeChatService struct {
	response chat.Response
	err      error
}

func (f fakeChatService) Send(context.Context, chat.Request) (chat.Response, error) {
	if f.err != nil {
		return chat.Response{}, f.err
	}
	return f.response, nil
}

func TestChatRouteRequiresAuth(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		ChatService: fakeChatService{},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestChatRouteReturnsAnswerAndUsage(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		ChatService: fakeChatService{
			response: chat.Response{
				Answer:         "Hello",
				ConversationID: "conversation-1",
				MessageUsage: chat.Usage{
					PromptTokens:     2,
					CompletionTokens: 3,
					TotalTokens:      5,
				},
				SessionTotalUsage: chat.Usage{
					PromptTokens:     2,
					CompletionTokens: 3,
					TotalTokens:      5,
				},
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"hello"}`))
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

func TestChatRouteMapsValidationError(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		ChatService: fakeChatService{err: chat.ErrMessageRequired},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":" "}`))
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

func TestChatRouteMapsUnexpectedServiceError(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
		ChatService: fakeChatService{err: errors.New("provider down")},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chat", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer session-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
}
