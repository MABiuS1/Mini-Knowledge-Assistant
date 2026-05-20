package httpapi

import (
	"context"
	"errors"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
	"github.com/mabius/knowledge-assistant/backend/internal/document"
)

type DocumentService interface {
	Upload(ctx context.Context, userID string, fileHeader *multipart.FileHeader) (document.Document, error)
}

type documentHandler struct {
	service DocumentService
}

func (h documentHandler) upload(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	doc, err := h.service.Upload(c.UserContext(), user.ID, fileHeader)
	if err != nil {
		return documentUploadError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"document": doc,
	})
}

func documentUploadError(err error) error {
	switch {
	case errors.Is(err, document.ErrFileRequired):
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	case errors.Is(err, document.ErrFileTooLarge):
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file is too large")
	case errors.Is(err, document.ErrInvalidType):
		return fiber.NewError(fiber.StatusBadRequest, "only PDF and TXT files are allowed")
	case errors.Is(err, document.ErrUnsafeName):
		return fiber.NewError(fiber.StatusBadRequest, "file name is not safe")
	default:
		return err
	}
}
