package httpapi

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mabius/knowledge-assistant/backend/internal/chat"
)

type ChatService interface {
	Send(ctx context.Context, req chat.Request) (chat.Response, error)
}

type chatHandler struct {
	service ChatService
}

type chatRequest struct {
	Message        string `json:"message"`
	ConversationID string `json:"conversationId"`
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
	})
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
