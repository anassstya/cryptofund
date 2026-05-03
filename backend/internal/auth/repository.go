package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type RepositoryAuth struct {
	db DB
}

func NewRepository(db DB) *RepositoryAuth {
	return &RepositoryAuth{db: db}
}

func (r *RepositoryAuth) AddUser(ctx context.Context, email, hash string) error {
	tag, err := r.db.Exec(ctx,
		`INSERT INTO users (email, password_hash) 
			VALUES ($1,  $2)
			ON CONFLICT (email) DO NOTHING`, email, hash)

	if err != nil {
		return fmt.Errorf("repository: add user: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrUserAlreadyExists
	}

	return nil
}

func (r *RepositoryAuth) GetUserID(ctx context.Context, email string) (string, error) {
	var userID string

	err := r.db.QueryRow(ctx,
		`SELECT id FROM users
				WHERE email=$1`, email).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", fmt.Errorf("failed to get user ID: %w", err)
	}

	return userID, nil
}

func (r *RepositoryAuth) GetPasswordHash(ctx context.Context, email string) (string, error) {
	var hash string

	err := r.db.QueryRow(ctx,
		`SELECT password_hash FROM users
				WHERE email=$1`, email).Scan(&hash)

	if err != nil {
		return "", fmt.Errorf("error to get password hash: %w", err)
	}

	return hash, nil
}
