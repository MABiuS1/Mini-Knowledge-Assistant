package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
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
	ID        string    `json:"id"`
	Role      Role      `json:"role"`
	Content   string    `json:"content"`
	Usage     Usage     `json:"usage"`
	CreatedAt time.Time `json:"createdAt"`
}

type Usage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

type Store interface {
	CreateConversation(ctx context.Context, userID string, title string) (string, error)
	VerifyConversationOwner(ctx context.Context, userID string, conversationID string) error
	ListConversations(ctx context.Context, userID string, limit int) ([]Conversation, error)
	ListMessages(ctx context.Context, conversationID string, limit int) ([]Message, error)
	CreateMessage(ctx context.Context, conversationID string, role Role, content string, usage Usage) (string, error)
	CreateUsageEvent(ctx context.Context, params UsageEventParams) error
	SumConversationUsage(ctx context.Context, conversationID string) (Usage, error)
}

type Conversation struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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

type Retriever interface {
	Retrieve(ctx context.Context, userID string, documentIDs []string, query string, limit int) ([]RetrievedChunk, error)
}

type Completion struct {
	Content string
	Model   string
	Usage   Usage
}

type Service struct {
	store     Store
	client    AIClient
	retriever Retriever
}

type Request struct {
	UserID         string
	ConversationID string
	Message        string
	DocumentIDs    []string
}

type Response struct {
	Answer             string     `json:"answer"`
	ConversationID     string     `json:"conversationId"`
	MessageUsage       Usage      `json:"usage"`
	SessionTotalUsage  Usage      `json:"sessionTotalUsage"`
	AssistantMessageID string     `json:"assistantMessageId"`
	Citations          []Citation `json:"citations"`
}

type ConversationDetail struct {
	Conversation      Conversation `json:"conversation"`
	Messages          []Message    `json:"messages"`
	SessionTotalUsage Usage        `json:"sessionTotalUsage"`
}

type RetrievedChunk struct {
	DocumentID string
	FileName   string
	ChunkID    string
	ChunkIndex int
	Content    string
	Similarity float64
}

type Citation struct {
	DocumentID string  `json:"documentId"`
	FileName   string  `json:"fileName"`
	ChunkID    string  `json:"chunkId"`
	ChunkIndex int     `json:"chunkIndex"`
	Snippet    string  `json:"snippet"`
	Similarity float64 `json:"similarity"`
}

func NewService(store Store, client AIClient, retriever Retriever) *Service {
	return &Service{
		store:     store,
		client:    client,
		retriever: retriever,
	}
}

func (s *Service) Send(ctx context.Context, req Request) (Response, error) {
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return Response{}, ErrMessageRequired
	}

	conversationID := strings.TrimSpace(req.ConversationID)
	if conversationID == "" {
		title := truncateRunes(message, 80)

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

	retrievedChunks := []RetrievedChunk{}
	if len(req.DocumentIDs) > 0 && s.retriever != nil {
		chunks, err := s.retriever.Retrieve(ctx, req.UserID, req.DocumentIDs, message, 5)
		if err != nil {
			return Response{}, err
		}
		retrievedChunks = chunks
	}

	history, err := s.store.ListMessages(ctx, conversationID, 20)
	if err != nil {
		return Response{}, err
	}

	messages := append(systemMessages(len(retrievedChunks) > 0), contextMessages(retrievedChunks)...)
	messages = append(messages, history...)

	completion, err := s.client.Complete(ctx, messages)
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
		Citations:          citationsFromChunks(retrievedChunks),
	}, nil
}

func (s *Service) ListConversations(ctx context.Context, userID string) ([]Conversation, error) {
	return s.store.ListConversations(ctx, userID, 50)
}

func (s *Service) LoadConversation(ctx context.Context, userID string, conversationID string) (ConversationDetail, error) {
	conversationID = strings.TrimSpace(conversationID)
	if conversationID == "" {
		return ConversationDetail{}, ErrConversationNotFound
	}

	if err := s.store.VerifyConversationOwner(ctx, userID, conversationID); err != nil {
		return ConversationDetail{}, ErrConversationNotFound
	}

	conversations, err := s.store.ListConversations(ctx, userID, 100)
	if err != nil {
		return ConversationDetail{}, err
	}

	var conversation Conversation
	for _, item := range conversations {
		if item.ID == conversationID {
			conversation = item
			break
		}
	}

	messages, err := s.store.ListMessages(ctx, conversationID, 200)
	if err != nil {
		return ConversationDetail{}, err
	}

	totalUsage, err := s.store.SumConversationUsage(ctx, conversationID)
	if err != nil {
		return ConversationDetail{}, err
	}

	if conversation.ID == "" {
		conversation.ID = conversationID
		conversation.Title = "Conversation"
	}

	return ConversationDetail{
		Conversation:      conversation,
		Messages:          messages,
		SessionTotalUsage: totalUsage,
	}, nil
}

func systemMessages(hasDocumentContext bool) []Message {
	content := "You are a concise knowledge assistant. Answer clearly. If the user asks for document-specific information and no document context is available, say that document context has not been provided yet."
	if hasDocumentContext {
		content = "You are a concise knowledge assistant. Use the provided document context to answer. If the context does not contain enough information, say so instead of guessing. When using document context, keep the answer grounded in that context."
	}

	return []Message{
		{
			Role:    RoleSystem,
			Content: content,
		},
	}
}

func contextMessages(chunks []RetrievedChunk) []Message {
	if len(chunks) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString("Document context:\n")
	for index, chunk := range chunks {
		builder.WriteString(fmt.Sprintf("[%d] %s chunk %d: %s\n", index+1, chunk.FileName, chunk.ChunkIndex, chunk.Content))
	}

	return []Message{{Role: RoleSystem, Content: builder.String()}}
}

func citationsFromChunks(chunks []RetrievedChunk) []Citation {
	citations := make([]Citation, 0, len(chunks))
	for _, chunk := range chunks {
		citations = append(citations, Citation{
			DocumentID: chunk.DocumentID,
			FileName:   chunk.FileName,
			ChunkID:    chunk.ChunkID,
			ChunkIndex: chunk.ChunkIndex,
			Snippet:    snippet(chunk.Content, 240),
			Similarity: chunk.Similarity,
		})
	}
	return citations
}

func snippet(content string, maxLength int) string {
	content = strings.TrimSpace(content)
	truncated := truncateRunes(content, maxLength)
	if truncated == content {
		return content
	}
	return strings.TrimSpace(truncated) + "..."
}

func truncateRunes(value string, maxLength int) string {
	if maxLength <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= maxLength {
		return value
	}
	return string(runes[:maxLength])
}
