package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

var ErrUserAlreadyExists = errors.New("user with this email already exists")

func (r *Repository) AddUser(ctx context.Context, email, hash string) error {
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

func (r *Repository) GetUser(ctx context.Context, email string) (string, error) {
	var userID string

	err := r.db.QueryRow(ctx,
		`SELECT id FROM users
				WHERE email=$1`,
		email).Scan(&userID)

	if err != nil {
		return "", fmt.Errorf("error to get user: %w", err)
	}

	return userID, nil
}
