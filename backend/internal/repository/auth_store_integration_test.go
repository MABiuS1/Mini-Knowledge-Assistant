package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/mabius/knowledge-assistant/backend/internal/auth"
)

func TestAuthStoreLoginWithSeededAdmin(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	db, err := Open(databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	service := auth.NewService(NewAuthStore(db), time.Hour)
	result, err := service.Login(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("login with seeded admin: %v", err)
	}

	if result.User.Username != "admin" {
		t.Fatalf("expected admin, got %q", result.User.Username)
	}
}
