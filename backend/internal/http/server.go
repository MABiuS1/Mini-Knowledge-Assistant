package httpapi

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mabius/knowledge-assistant/backend/internal/auth"
	"github.com/mabius/knowledge-assistant/backend/internal/config"
	"github.com/mabius/knowledge-assistant/backend/internal/repository"
)

func NewServer(cfg config.Config) *fiber.App {
	db, err := repository.Open(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	authStore := repository.NewAuthStore(db)
	authService := authAdapter{service: auth.NewService(authStore, cfg.SessionTTL)}

	return NewServerWithDependencies(cfg, Dependencies{
		AuthService: authService,
	})
}

type Dependencies struct {
	AuthService AuthService
}

func NewServerWithDependencies(cfg config.Config, deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Knowledge Assistant API",
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
