package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUnauthenticated    = errors.New("unauthenticated")
)

type User struct {
	ID           string
	Username     string
	PasswordHash string
}

type Store interface {
	FindUserByUsername(ctx context.Context, username string) (User, error)
	FindUserBySessionHash(ctx context.Context, tokenHash string, now time.Time) (User, error)
	CreateSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, tokenHash string) error
}

type Service struct {
	store      Store
	sessionTTL time.Duration
}

type LoginResult struct {
	User      User
	Token     string
	ExpiresAt time.Time
}

func NewService(store Store, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = 24 * time.Hour
	}

	return &Service{
		store:      store,
		sessionTTL: sessionTTL,
	}
}

func (s *Service) Login(ctx context.Context, username string, password string) (LoginResult, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return LoginResult{}, ErrInvalidCredentials
	}

	user, err := s.store.FindUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return LoginResult{}, err
	}

	expiresAt := time.Now().UTC().Add(s.sessionTTL)
	if err := s.store.CreateSession(ctx, user.ID, HashToken(token), expiresAt); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		User:      user,
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return User{}, ErrUnauthenticated
	}

	user, err := s.store.FindUserBySessionHash(ctx, HashToken(token), time.Now().UTC())
	if err != nil {
		return User{}, ErrUnauthenticated
	}

	return user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil
	}

	return s.store.DeleteSession(ctx, HashToken(token))
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func generateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}
