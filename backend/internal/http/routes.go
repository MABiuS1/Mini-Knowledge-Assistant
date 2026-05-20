package httpapi

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mabius/knowledge-assistant/backend/internal/config"
)

func registerRoutes(app *fiber.App, cfg config.Config, deps Dependencies) {
	api := app.Group("/api")

	api.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.AppEnv,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
	})

	authHandler := authHandler{cfg: cfg, service: deps.AuthService}
	api.Post("/auth/login", authHandler.login)
	api.Post("/auth/logout", authHandler.logout)

	protected := api.Group("", authMiddleware(deps.AuthService))
	protected.Get("/me", authHandler.me)

	documentHandler := documentHandler{service: deps.DocumentService}
	protected.Post("/documents/upload", documentHandler.upload)

	chatHandler := chatHandler{service: deps.ChatService}
	protected.Post("/chat", chatHandler.send)
}
