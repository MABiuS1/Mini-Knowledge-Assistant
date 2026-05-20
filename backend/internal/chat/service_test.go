package chat

import (
	"context"
	"errors"
	"testing"
	"unicode/utf8"
)

type memoryStore struct {
	conversationID string
	lastTitle      string
	messages       []Message
	usage          Usage
}

func (m *memoryStore) CreateConversation(_ context.Context, _ string, title string) (string, error) {
	m.conversationID = "conversation-1"
	m.lastTitle = title
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
	}, nil)

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
	service := NewService(&memoryStore{}, fakeClient{}, nil)

	_, err := service.Send(context.Background(), Request{UserID: "user-1", Message: " \n\t "})
	if !errors.Is(err, ErrMessageRequired) {
		t.Fatalf("expected message required, got %v", err)
	}
}

func TestSendRejectsEmptyAssistantResponse(t *testing.T) {
	service := NewService(&memoryStore{}, fakeClient{
		completion: Completion{Content: "  "},
	}, nil)

	_, err := service.Send(context.Background(), Request{UserID: "user-1", Message: "Hello"})
	if !errors.Is(err, ErrAssistantEmptyMessage) {
		t.Fatalf("expected empty assistant error, got %v", err)
	}
}

func TestSendTruncatesThaiConversationTitleAsUTF8(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, fakeClient{
		completion: Completion{Content: "ตอบกลับ"},
	}, nil)

	_, err := service.Send(context.Background(), Request{
		UserID:  "user-1",
		Message: "สรุปประสบการณ์ทำงานจากเอกสารนี้และช่วยอธิบายทักษะเด่นทั้งหมดโดยละเอียดมากพอสำหรับอ่านก่อนสัมภาษณ์",
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if !utf8.ValidString(store.lastTitle) {
		t.Fatalf("expected valid utf-8 title, got %q", store.lastTitle)
	}

	if got := len([]rune(store.lastTitle)); got != 80 {
		t.Fatalf("expected title to be truncated to 80 runes, got %d", got)
	}
}

type fakeRetriever struct {
	chunks []RetrievedChunk
	err    error
}

func (f fakeRetriever) Retrieve(context.Context, string, []string, string, int) ([]RetrievedChunk, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.chunks, nil
}

func TestSendWithDocumentContextReturnsCitations(t *testing.T) {
	store := &memoryStore{}
	service := NewService(store, fakeClient{
		completion: Completion{
			Content: "The document says hello.",
			Model:   "test-model",
			Usage:   Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		},
	}, fakeRetriever{
		chunks: []RetrievedChunk{
			{
				DocumentID: "document-1",
				FileName:   "notes.txt",
				ChunkID:    "chunk-1",
				ChunkIndex: 0,
				Content:    "hello from the uploaded document",
				Similarity: 0.91,
			},
		},
	})

	response, err := service.Send(context.Background(), Request{
		UserID:      "user-1",
		Message:     "What does the document say?",
		DocumentIDs: []string{"document-1"},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	if len(response.Citations) != 1 {
		t.Fatalf("expected one citation, got %d", len(response.Citations))
	}

	if response.Citations[0].FileName != "notes.txt" {
		t.Fatalf("expected citation file name, got %q", response.Citations[0].FileName)
	}
}

func TestSnippetPreservesUTF8(t *testing.T) {
	got := snippet("สวัสดีครับนี่คือข้อความภาษาไทย", 7)
	want := "สวัสดีค..."

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	if !utf8.ValidString(got) {
		t.Fatalf("expected valid utf-8, got %q", got)
	}
}
