package chat

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrMessageRequired       = errors.New("message is required")
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrAssistantEmptyMessage = errors.New("assistant returned an empty message")
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

type Message struct {
	ID      string
	Role    Role
	Content string
	Usage   Usage
}

type Usage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

type Store interface {
	CreateConversation(ctx context.Context, userID string, title string) (string, error)
	VerifyConversationOwner(ctx context.Context, userID string, conversationID string) error
	ListMessages(ctx context.Context, conversationID string, limit int) ([]Message, error)
	CreateMessage(ctx context.Context, conversationID string, role Role, content string, usage Usage) (string, error)
	CreateUsageEvent(ctx context.Context, params UsageEventParams) error
	SumConversationUsage(ctx context.Context, conversationID string) (Usage, error)
}

type UsageEventParams struct {
	UserID         string
	ConversationID string
	MessageID      string
	Provider       string
	Model          string
	Usage          Usage
}

type AIClient interface {
	Complete(ctx context.Context, messages []Message) (Completion, error)
}

type Completion struct {
	Content string
	Model   string
	Usage   Usage
}

type Service struct {
	store  Store
	client AIClient
}

type Request struct {
	UserID         string
	ConversationID string
	Message        string
}

type Response struct {
	Answer             string `json:"answer"`
	ConversationID     string `json:"conversationId"`
	MessageUsage       Usage  `json:"usage"`
	SessionTotalUsage  Usage  `json:"sessionTotalUsage"`
	AssistantMessageID string `json:"assistantMessageId"`
}

func NewService(store Store, client AIClient) *Service {
	return &Service{
		store:  store,
		client: client,
	}
}

func (s *Service) Send(ctx context.Context, req Request) (Response, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return Response{}, ErrMessageRequired
	}

	conversationID := strings.TrimSpace(req.ConversationID)
	if conversationID == "" {
		title := message
		if len(title) > 80 {
			title = title[:80]
		}

		createdID, err := s.store.CreateConversation(ctx, req.UserID, title)
		if err != nil {
			return Response{}, err
		}
		conversationID = createdID
	} else if err := s.store.VerifyConversationOwner(ctx, req.UserID, conversationID); err != nil {
		return Response{}, ErrConversationNotFound
	}

	if _, err := s.store.CreateMessage(ctx, conversationID, RoleUser, message, Usage{}); err != nil {
		return Response{}, err
	}

	history, err := s.store.ListMessages(ctx, conversationID, 20)
	if err != nil {
		return Response{}, err
	}

	completion, err := s.client.Complete(ctx, append(systemMessages(), history...))
	if err != nil {
		return Response{}, err
	}

	answer := strings.TrimSpace(completion.Content)
	if answer == "" {
		return Response{}, ErrAssistantEmptyMessage
	}

	assistantMessageID, err := s.store.CreateMessage(ctx, conversationID, RoleAssistant, answer, completion.Usage)
	if err != nil {
		return Response{}, err
	}

	if err := s.store.CreateUsageEvent(ctx, UsageEventParams{
		UserID:         req.UserID,
		ConversationID: conversationID,
		MessageID:      assistantMessageID,
		Provider:       "openai",
		Model:          completion.Model,
		Usage:          completion.Usage,
	}); err != nil {
		return Response{}, err
	}

	totalUsage, err := s.store.SumConversationUsage(ctx, conversationID)
	if err != nil {
		return Response{}, err
	}

	return Response{
		Answer:             answer,
		ConversationID:     conversationID,
		MessageUsage:       completion.Usage,
		SessionTotalUsage:  totalUsage,
		AssistantMessageID: assistantMessageID,
	}, nil
}

func systemMessages() []Message {
	return []Message{
		{
			Role:    RoleSystem,
			Content: "You are a concise knowledge assistant. Answer clearly. If the user asks for document-specific information and no document context is available, say that document context has not been provided yet.",
		},
	}
}
