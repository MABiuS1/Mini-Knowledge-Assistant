package httpapi

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mabius/knowledge-assistant/backend/internal/rag"
)

type RAGService interface {
	Retrieve(ctx context.Context, userID string, documentIDs []string, query string, limit int) ([]rag.RetrievedChunk, error)
}

type ragHandler struct {
	service RAGService
}

type retrieveRequest struct {
	Query       string   `json:"query"`
	DocumentIDs []string `json:"documentIds"`
	Limit       int      `json:"limit"`
}

func (h ragHandler) retrieve(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	var req retrieveRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "query is required")
	}

	chunks, err := h.service.Retrieve(c.UserContext(), user.ID, req.DocumentIDs, query, req.Limit)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"chunks": chunks,
	})
}
