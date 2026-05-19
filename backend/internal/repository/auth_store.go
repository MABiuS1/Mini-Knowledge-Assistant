package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mabius/knowledge-assistant/backend/internal/auth"
)

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) *AuthStore {
	return &AuthStore{db: db}
}

func (s *AuthStore) FindUserByUsername(ctx context.Context, username string) (auth.User, error) {
	var user auth.User
	err := s.db.QueryRowContext(ctx, `
		SELECT id::text, username, password_hash
		FROM users
		WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrInvalidCredentials
	}
	return user, err
}

func (s *AuthStore) FindUserBySessionHash(ctx context.Context, tokenHash string, now time.Time) (auth.User, error) {
	var user auth.User
	err := s.db.QueryRowContext(ctx, `
		SELECT users.id::text, users.username, users.password_hash
		FROM sessions
		INNER JOIN users ON users.id = sessions.user_id
		WHERE sessions.token_hash = $1
		  AND sessions.expires_at > $2
	`, tokenHash, now).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if errors.Is(err, sql.ErrNoRows) {
		return auth.User{}, auth.ErrUnauthenticated
	}
	return user, err
}

func (s *AuthStore) CreateSession(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1::uuid, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (s *AuthStore) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := s.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE token_hash = $1
	`, tokenHash)
	return err
}
