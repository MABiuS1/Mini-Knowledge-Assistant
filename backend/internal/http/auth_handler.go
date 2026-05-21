package httpapi

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mabius/knowledge-assistant/backend/internal/auth"
	"github.com/mabius/knowledge-assistant/backend/internal/config"
)

const sessionCookieName = "session_token"

type AuthService interface {
	Login(ctx context.Context, username string, password string) (auth.LoginResult, error)
	Authenticate(ctx context.Context, token string) (auth.User, error)
	Logout(ctx context.Context, token string) error
}

type authAdapter struct {
	service *auth.Service
}

func (a authAdapter) Login(ctx context.Context, username string, password string) (auth.LoginResult, error) {
	return a.service.Login(ctx, username, password)
}

func (a authAdapter) Authenticate(ctx context.Context, token string) (auth.User, error) {
	return a.service.Authenticate(ctx, token)
}

func (a authAdapter) Logout(ctx context.Context, token string) error {
	return a.service.Logout(ctx, token)
}

type authHandler struct {
	cfg     config.Config
	service AuthService
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h authHandler) login(c *fiber.Ctx) error {
	var req loginRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	result, err := h.service.Login(c.UserContext(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid username or password")
		}
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    result.Token,
		Expires:  result.ExpiresAt,
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: sessionCookieSameSite(h.cfg),
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"user": publicUser(result.User),
	})
}

func (h authHandler) logout(c *fiber.Ctx) error {
	token := tokenFromRequest(c)
	if err := h.service.Logout(c.UserContext(), token); err != nil {
		return err
	}

	clearSessionCookie(c, h.cfg)
	return c.JSON(fiber.Map{"ok": true})
}

func (h authHandler) me(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}

	return c.JSON(fiber.Map{
		"user": publicUser(user),
	})
}

func authMiddleware(service AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := tokenFromRequest(c)
		user, err := service.Authenticate(c.UserContext(), token)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
		}

		c.Locals("user", user)
		return c.Next()
	}
}

func tokenFromRequest(c *fiber.Ctx) string {
	if token := c.Cookies(sessionCookieName); token != "" {
		return token
	}

	header := c.Get("Authorization")
	if header == "" {
		return ""
	}

	value, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return ""
	}

	return strings.TrimSpace(value)
}

func currentUser(c *fiber.Ctx) (auth.User, bool) {
	user, ok := c.Locals("user").(auth.User)
	return user, ok
}

func publicUser(user auth.User) fiber.Map {
	return fiber.Map{
		"id":       user.ID,
		"username": user.Username,
	}
}

func clearSessionCookie(c *fiber.Ctx, cfg config.Config) {
	c.Cookie(&fiber.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		HTTPOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sessionCookieSameSite(cfg),
		Path:     "/",
		MaxAge:   -1,
	})
}

func sessionCookieSameSite(cfg config.Config) string {
	if cfg.CookieSecure {
		return "None"
	}

	return "Lax"
}
