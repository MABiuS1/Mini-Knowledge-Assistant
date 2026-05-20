package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type memoryStore struct {
	user     User
	sessions map[string]time.Time
}

func newMemoryStore(t *testing.T) *memoryStore {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	return &memoryStore{
		user: User{
			ID:           "user-1",
			Username:     "admin",
			PasswordHash: string(hash),
		},
		sessions: map[string]time.Time{},
	}
}

func (m *memoryStore) FindUserByUsername(_ context.Context, username string) (User, error) {
	if username != m.user.Username {
		return User{}, errors.New("not found")
	}
	return m.user, nil
}

func (m *memoryStore) FindUserBySessionHash(_ context.Context, tokenHash string, now time.Time) (User, error) {
	expiresAt, ok := m.sessions[tokenHash]
	if !ok || !expiresAt.After(now) {
		return User{}, errors.New("not found")
	}
	return m.user, nil
}

func (m *memoryStore) CreateSession(_ context.Context, _ string, tokenHash string, expiresAt time.Time) error {
	m.sessions[tokenHash] = expiresAt
	return nil
}

func (m *memoryStore) DeleteSession(_ context.Context, tokenHash string) error {
	delete(m.sessions, tokenHash)
	return nil
}

func TestLoginSuccess(t *testing.T) {
	store := newMemoryStore(t)
	service := NewService(store, time.Hour)

	result, err := service.Login(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if result.Token == "" {
		t.Fatal("expected session token")
	}

	if result.User.Username != "admin" {
		t.Fatalf("expected admin user, got %q", result.User.Username)
	}
}

func TestSeededAdminHashMatchesDefaultPassword(t *testing.T) {
	const seedHash = "$2a$10$V6NeE.IlX99qdan7ORKcfe1pZOxUI0pHcPcb6nI3sXlSjBa4RKPda"

	if err := bcrypt.CompareHashAndPassword([]byte(seedHash), []byte("admin123")); err != nil {
		t.Fatalf("seeded admin hash should match admin123: %v", err)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	store := newMemoryStore(t)
	service := NewService(store, time.Hour)

	_, err := service.Login(context.Background(), "admin", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}

func TestLoginReturnsStoreError(t *testing.T) {
	storeErr := errors.New("database unavailable")
	service := NewService(failingStore{err: storeErr}, time.Hour)

	_, err := service.Login(context.Background(), "admin", "admin123")
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected store error, got %v", err)
	}
}

func TestAuthenticateRequiresValidSession(t *testing.T) {
	store := newMemoryStore(t)
	service := NewService(store, time.Hour)

	result, err := service.Login(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	user, err := service.Authenticate(context.Background(), result.Token)
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}

	if user.ID != "user-1" {
		t.Fatalf("expected user-1, got %q", user.ID)
	}
}

func TestAuthenticateRejectsEmptyToken(t *testing.T) {
	service := NewService(newMemoryStore(t), time.Hour)

	_, err := service.Authenticate(context.Background(), " \n\t ")
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestLogoutDeletesSession(t *testing.T) {
	store := newMemoryStore(t)
	service := NewService(store, time.Hour)

	result, err := service.Login(context.Background(), "admin", "admin123")
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if err := service.Logout(context.Background(), result.Token); err != nil {
		t.Fatalf("logout: %v", err)
	}

	_, err = service.Authenticate(context.Background(), result.Token)
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("expected unauthenticated after logout, got %v", err)
	}
}

type failingStore struct {
	err error
}

func (f failingStore) FindUserByUsername(context.Context, string) (User, error) {
	return User{}, f.err
}

func (f failingStore) FindUserBySessionHash(context.Context, string, time.Time) (User, error) {
	return User{}, f.err
}

func (f failingStore) CreateSession(context.Context, string, string, time.Time) error {
	return f.err
}

func (f failingStore) DeleteSession(context.Context, string) error {
	return f.err
}
