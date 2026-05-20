package chat

import (
	"context"
	"errors"
	"testing"
)

type memoryStore struct {
	conversationID string
	messages       []Message
	usage          Usage
}

func (m *memoryStore) CreateConversation(context.Context, string, string) (string, error) {
	m.conversationID = "conversation-1"
	return m.conversationID, nil
}

func (m *memoryStore) VerifyConversationOwner(context.Context, string, string) error {
	if m.conversationID == "" {
		return errors.New("not found")
	}
	return nil
}

func (m *memoryStore) ListMessages(context.Context, string, int) ([]Message, error) {
	return m.messages, nil
}

func (m *memoryStore) CreateMessage(_ context.Context, _ string, role Role, content string, usage Usage) (string, error) {
	id := "message-1"
	if role == RoleAssistant {
		id = "assistant-message-1"
	}

	m.messages = append(m.messages, Message{
		ID:      id,
		Role:    role,
		Content: content,
		Usage:   usage,
	})

	if role == RoleAssistant {
		m.usage.PromptTokens += usage.PromptTokens
		m.usage.CompletionTokens += usage.CompletionTokens
		m.usage.TotalTokens += usage.TotalTokens
	}

	return id, nil
}

func (m *memoryStore) CreateUsageEvent(context.Context, UsageEventParams) error {
	return nil
}

func (m *memoryStore) SumConversationUsage(context.Context, string) (Usage, error) {
	return m.usage, nil
}

type fakeClient struct {
	completion Completion
	err        error
}

func (f fakeClient) Complete(context.Context, []Message) (Completion, error) {
	if f.err != nil {
		return Completion{}, f.err
	}
	return f.completion, nil
}

func TestSendCreatesConversationAndTracksUsage(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, fakeClient{
		completion: Completion{
			Content: "Hello from AI",
			Model:   "test-model",
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 4,
				TotalTokens:      14,
			},
		},
	})

	response, err := service.Send(context.Background(), Request{
		UserID:  "user-1",
		Message: "Hello",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if response.ConversationID != "conversation-1" {
		t.Fatalf("expected conversation id, got %q", response.ConversationID)
	}

	if response.Answer != "Hello from AI" {
		t.Fatalf("expected answer, got %q", response.Answer)
	}

	if response.MessageUsage.TotalTokens != 14 {
		t.Fatalf("expected message tokens, got %d", response.MessageUsage.TotalTokens)
	}

	if response.SessionTotalUsage.TotalTokens != 14 {
		t.Fatalf("expected session total tokens, got %d", response.SessionTotalUsage.TotalTokens)
	}
}

func TestSendRejectsEmptyMessage(t *testing.T) {
	service := NewService(&memoryStore{}, fakeClient{})

	_, err := service.Send(context.Background(), Request{UserID: "user-1", Message: " \n\t "})
	if !errors.Is(err, ErrMessageRequired) {
		t.Fatalf("expected message required, got %v", err)
	}
}

func TestSendRejectsEmptyAssistantResponse(t *testing.T) {
	service := NewService(&memoryStore{}, fakeClient{
		completion: Completion{Content: "  "},
	})

	_, err := service.Send(context.Background(), Request{UserID: "user-1", Message: "Hello"})
	if !errors.Is(err, ErrAssistantEmptyMessage) {
		t.Fatalf("expected empty assistant error, got %v", err)
	}
}
