package httpapi

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mabius/knowledge-assistant/backend/internal/auth"
	"github.com/mabius/knowledge-assistant/backend/internal/chat"
	"github.com/mabius/knowledge-assistant/backend/internal/config"
	"github.com/mabius/knowledge-assistant/backend/internal/document"
	"github.com/mabius/knowledge-assistant/backend/internal/repository"
)

func NewServer(cfg config.Config) *fiber.App {
	db, err := repository.Open(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	authStore := repository.NewAuthStore(db)
	authService := authAdapter{service: auth.NewService(authStore, cfg.SessionTTL)}
	documentStore := repository.NewDocumentStore(db)
	documentService := document.NewService(documentStore, cfg.UploadDir, cfg.MaxUploadBytes)
	chatStore := repository.NewChatStore(db)
	aiClient, err := newAIClient(cfg)
	if err != nil {
		panic(err)
	}
	chatService := chat.NewService(chatStore, aiClient)

	return NewServerWithDependencies(cfg, Dependencies{
		AuthService:     authService,
		DocumentService: documentService,
		ChatService:     chatService,
	})
}

type Dependencies struct {
	AuthService     AuthService
	DocumentService DocumentService
	ChatService     ChatService
}

func NewServerWithDependencies(cfg config.Config, deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Knowledge Assistant API",
		BodyLimit:    int(cfg.MaxUploadBytes + 1024*1024),
		ReadTimeout:  cfg.RequestTimeout,
		WriteTimeout: cfg.RequestTimeout,
		IdleTimeout:  60 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if fiberErr, ok := err.(*fiber.Error); ok {
				code = fiberErr.Code
			}

			return c.Status(code).JSON(fiber.Map{
				"error": fiber.Map{
					"message": err.Error(),
				},
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL,
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
	}))

	registerRoutes(app, cfg, deps)

	return app
}

func newAIClient(cfg config.Config) (chat.AIClient, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.AIProvider)) {
	case "openai":
		return chat.NewOpenAIClient(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, cfg.RequestTimeout), nil
	case "gemini":
		return chat.NewGeminiClient(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, cfg.RequestTimeout), nil
	default:
		return nil, fmt.Errorf("unsupported AI_PROVIDER %q", cfg.AIProvider)
	}
}
