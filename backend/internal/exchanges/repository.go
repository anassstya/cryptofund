package exchanges

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
}

type RepositoryExchanges struct {
	db DB
}

func NewRepository(db DB) *RepositoryExchanges {
	return &RepositoryExchanges{db: db}
}

func (r *RepositoryExchanges) AddExchange(ctx context.Context, userID, name, KeyAPI, SecretAPI string) (string, error) {
	var newID string
	err := r.db.QueryRow(ctx,
		`INSERT INTO exchanges (user_id, name, api_key, api_secret)
                VALUES ($1, $2, $3, $4) RETURNING id`, userID, name, KeyAPI, SecretAPI).Scan(&newID)

	if err != nil {
		return "", fmt.Errorf("repository: adding error %w", err)
	}
	return newID, nil
}

func (r *RepositoryExchanges) GetByUserID(ctx context.Context, userID string) ([]Exchange, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, api_key, api_secret FROM exchanges
		 WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: query exchanges: %w", err)
	}
	defer rows.Close()

	var exchanges []Exchange
	for rows.Next() {
		var ex Exchange
		err := rows.Scan(&ex.ID, &ex.UserID, &ex.Name, &ex.KeyAPI, &ex.SecretAPI)
		if err != nil {
			return nil, fmt.Errorf("repository: scan exchange: %w", err)
		}
		exchanges = append(exchanges, ex)
	}
	return exchanges, rows.Err()
}
