package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mabius/knowledge-assistant/backend/internal/auth"
	"github.com/mabius/knowledge-assistant/backend/internal/config"
)

type fakeAuthService struct {
	loginErr        error
	authenticateErr error
	logoutCalled    bool
}

func (f *fakeAuthService) Login(_ context.Context, username string, password string) (auth.LoginResult, error) {
	if f.loginErr != nil {
		return auth.LoginResult{}, f.loginErr
	}

	if username != "admin" || password != "admin123" {
		return auth.LoginResult{}, auth.ErrInvalidCredentials
	}

	return auth.LoginResult{
		User: auth.User{
			ID:       "user-1",
			Username: "admin",
		},
		Token:     "session-token",
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}, nil
}

func (f *fakeAuthService) Authenticate(_ context.Context, token string) (auth.User, error) {
	if f.authenticateErr != nil {
		return auth.User{}, f.authenticateErr
	}

	if token != "session-token" {
		return auth.User{}, auth.ErrUnauthenticated
	}

	return auth.User{ID: "user-1", Username: "admin"}, nil
}

func (f *fakeAuthService) Logout(_ context.Context, _ string) error {
	f.logoutCalled = true
	return nil
}

func TestLoginRouteSuccess(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
	})

	body := bytes.NewBufferString(`{"username":"admin","password":"admin123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if cookie := resp.Header.Get("Set-Cookie"); !strings.Contains(cookie, sessionCookieName+"=session-token") {
		t.Fatalf("expected session cookie, got %q", cookie)
	}
}

func TestLoginRouteInvalidCredentials(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
	})

	body := bytes.NewBufferString(`{"username":"admin","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMeRouteReturnsCurrentUser(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer session-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload struct {
		User struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"user"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload.User.Username != "admin" {
		t.Fatalf("expected admin, got %q", payload.User.Username)
	}
}

func TestLogoutRouteClearsSession(t *testing.T) {
	authService := &fakeAuthService{}
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: authService,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer session-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if !authService.logoutCalled {
		t.Fatal("expected logout to be called")
	}

	if cookie := resp.Header.Get("Set-Cookie"); !strings.Contains(cookie, sessionCookieName+"=;") {
		t.Fatalf("expected clearing cookie, got %q", cookie)
	}
}

func TestAuthMiddlewareMapsServiceErrorToUnauthorized(t *testing.T) {
	app := NewServerWithDependencies(testConfig(), Dependencies{
		AuthService: &fakeAuthService{authenticateErr: errors.New("store down")},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer session-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func testConfig() config.Config {
	return config.Config{
		AppEnv:         "test",
		Port:           "8080",
		FrontendURL:    "http://localhost:3000",
		RequestTimeout: time.Second,
		SessionTTL:     time.Hour,
	}
}
