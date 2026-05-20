package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mabius/knowledge-assistant/backend/internal/chat"
)

type ChatService interface {
	Send(ctx context.Context, req chat.Request) (chat.Response, error)
	ListConversations(ctx context.Context, userID string) ([]chat.Conversation, error)
	LoadConversation(ctx context.Context, userID string, conversationID string) (chat.ConversationDetail, error)
}

type chatHandler struct {
	service ChatService
}

type chatRequest struct {
	Message        string   `json:"message"`
	ConversationID string   `json:"conversationId"`
	DocumentIDs    []string `json:"documentIds"`
}

func (h chatHandler) send(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	var req chatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	response, err := h.service.Send(c.UserContext(), chat.Request{
		UserID:         user.ID,
		ConversationID: strings.TrimSpace(req.ConversationID),
		Message:        req.Message,
		DocumentIDs:    req.DocumentIDs,
	})
	if err != nil {
		return chatSendError(err)
	}

	return c.JSON(response)
}

func (h chatHandler) stream(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	var req chatRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	response, err := h.service.Send(c.UserContext(), chat.Request{
		UserID:         user.ID,
		ConversationID: strings.TrimSpace(req.ConversationID),
		Message:        req.Message,
		DocumentIDs:    req.DocumentIDs,
	})
	if err != nil {
		return chatSendError(err)
	}

	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, "keep-alive")
	c.Context().SetBodyStreamWriter(func(writer *bufio.Writer) {
		for _, chunk := range answerChunks(response.Answer, 28) {
			if err := writeSSE(writer, "delta", fiber.Map{"content": chunk}); err != nil {
				return
			}
			if err := writer.Flush(); err != nil {
				return
			}
			time.Sleep(15 * time.Millisecond)
		}

		_ = writeSSE(writer, "done", response)
		_ = writer.Flush()
	})

	return nil
}

func (h chatHandler) listConversations(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	conversations, err := h.service.ListConversations(c.UserContext(), user.ID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"conversations": conversations,
	})
}

func (h chatHandler) loadConversation(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	response, err := h.service.LoadConversation(c.UserContext(), user.ID, c.Params("id"))
	if err != nil {
		return chatSendError(err)
	}

	return c.JSON(response)
}

func chatSendError(err error) error {
	switch {
	case errors.Is(err, chat.ErrMessageRequired):
		return fiber.NewError(fiber.StatusBadRequest, "message is required")
	case errors.Is(err, chat.ErrConversationNotFound):
		return fiber.NewError(fiber.StatusNotFound, "conversation not found")
	case errors.Is(err, chat.ErrAssistantEmptyMessage):
		return fiber.NewError(fiber.StatusBadGateway, "assistant returned an empty message")
	default:
		return err
	}
}

func writeSSE(writer *bufio.Writer, event string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(writer, "event: %s\n", event); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "data: %s\n\n", data); err != nil {
		return err
	}

	return nil
}

func answerChunks(answer string, maxRunes int) []string {
	if maxRunes <= 0 {
		maxRunes = 28
	}

	runes := []rune(answer)
	if len(runes) == 0 {
		return []string{}
	}

	chunks := []string{}
	for start := 0; start < len(runes); start += maxRunes {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}

	return chunks
}
