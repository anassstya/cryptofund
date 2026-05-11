package exchanges

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
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

func (r *RepositoryExchanges) AddExchange(
	ctx context.Context,
	userID, name, keyAPI, secretAPI, credentialsHash string,
) (string, error) {
	var newID string

	err := r.db.QueryRow(ctx,
		`INSERT INTO exchanges (user_id, name, api_key, api_secret, credentials_hash)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		userID,
		name,
		keyAPI,
		secretAPI,
		credentialsHash,
	).Scan(&newID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return "", fmt.Errorf("repository: %w", ErrExchangeAlreadyExists)
		}

		return "", fmt.Errorf("repository: adding error: %w", err)
	}

	return newID, nil
}

func (r *RepositoryExchanges) GetByUserID(ctx context.Context, userID string) ([]ExchangeCreateResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("repository: user id is empty")
	}

	rows, err := r.db.Query(ctx,
		`SELECT id, name
		 FROM exchanges
		 WHERE user_id = $1
		 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: query exchanges: %w", err)
	}
	defer rows.Close()

	var exchanges []ExchangeCreateResponse

	for rows.Next() {
		var ex ExchangeCreateResponse

		err := rows.Scan(&ex.ID, &ex.Name)
		if err != nil {
			return nil, fmt.Errorf("repository: scan exchange: %w", err)
		}

		exchanges = append(exchanges, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows exchanges: %w", err)
	}

	return exchanges, nil
}

func (r *RepositoryExchanges) AddBalanceByExchangeID(
	ctx context.Context,
	id string,
	balance float64,
	changePercent float64,
	assetsCount int,
	source string,
) error {
	if id == "" {
		return fmt.Errorf("repository: exchange id is empty")
	}

	if source == "" {
		source = "mock"
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO exchange_balance (
			exchange_id,
			total_balance,
			change_percent,
			assets_count,
			source
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (exchange_id) DO UPDATE SET
			total_balance = EXCLUDED.total_balance,
			change_percent = EXCLUDED.change_percent,
			assets_count = EXCLUDED.assets_count,
			source = EXCLUDED.source,
			updated_at = NOW()`,
		id,
		balance,
		changePercent,
		assetsCount,
		source,
	)

	if err != nil {
		return fmt.Errorf("repository: add exchange balance: %w", err)
	}

	return nil
}

func (r *RepositoryExchanges) GetBalanceByExchangeID(ctx context.Context, id string) (ExchangeBalanceResponse, error) {
	if id == "" {
		return ExchangeBalanceResponse{}, fmt.Errorf("repository: exchange id is empty")
	}

	var res ExchangeBalanceResponse

	err := r.db.QueryRow(ctx,
		`SELECT 
			ex.id,
			ex.name,
			eb.total_balance,
			eb.change_percent,
			eb.assets_count,
			eb.source,
			eb.updated_at
		FROM exchange_balance eb
		JOIN exchanges ex ON eb.exchange_id = ex.id
		WHERE eb.exchange_id = $1`,
		id,
	).Scan(
		&res.ID,
		&res.Name,
		&res.Balance,
		&res.ChangePercent,
		&res.AssetsCount,
		&res.Source,
		&res.UpdatedAt,
	)

	if err != nil {
		return ExchangeBalanceResponse{}, fmt.Errorf("repository: get balance: %w", err)
	}

	return res, nil
}

func (r *RepositoryExchanges) GetMockBalancesByUserID(ctx context.Context, userID string) ([]ExchangeBalanceResponse, error) {
	if userID == "" {
		return nil, fmt.Errorf("repository: user id is empty")
	}

	rows, err := r.db.Query(ctx,
		`SELECT 
			ex.id,
			ex.name,
			eb.total_balance,
			eb.change_percent,
			eb.assets_count,
			eb.source,
			eb.updated_at
		FROM exchange_balance eb
		JOIN exchanges ex ON eb.exchange_id = ex.id
		WHERE ex.user_id = $1
		  AND eb.source = 'mock'
		ORDER BY eb.updated_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: query user mock balances: %w", err)
	}
	defer rows.Close()

	var balances []ExchangeBalanceResponse

	for rows.Next() {
		var balance ExchangeBalanceResponse

		err := rows.Scan(
			&balance.ID,
			&balance.Name,
			&balance.Balance,
			&balance.ChangePercent,
			&balance.AssetsCount,
			&balance.Source,
			&balance.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: scan user mock balance: %w", err)
		}

		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows user mock balances: %w", err)
	}

	return balances, nil
}

func (r *RepositoryExchanges) GetExchangesForUpdateByUserID(ctx context.Context, userID string) ([]ExchangeForUpdate, error) {
	if userID == "" {
		return nil, fmt.Errorf("repository: user id is empty")
	}

	rows, err := r.db.Query(ctx,
		`SELECT 
			ex.id,
			ex.name,
			ex.api_key,
			ex.api_secret,
			eb.total_balance,
			eb.assets_count,
			eb.source
		FROM exchanges ex
		JOIN exchange_balance eb ON eb.exchange_id = ex.id
		WHERE ex.user_id = $1
		ORDER BY eb.updated_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: query exchanges for update: %w", err)
	}
	defer rows.Close()

	var exchanges []ExchangeForUpdate

	for rows.Next() {
		var ex ExchangeForUpdate

		err := rows.Scan(
			&ex.ID,
			&ex.Name,
			&ex.KeyAPI,
			&ex.SecretAPI,
			&ex.Balance,
			&ex.AssetsCount,
			&ex.Source,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: scan exchange for update: %w", err)
		}

		exchanges = append(exchanges, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: rows exchanges for update: %w", err)
	}

	return exchanges, nil
}

func (r *RepositoryExchanges) AddBalanceHistory(ctx context.Context, exchangeID string, balance float64) error {
	if exchangeID == "" {
		return fmt.Errorf("repository: exchange id is empty")
	}

	_, err := r.db.Exec(ctx,
		`INSERT INTO exchange_balance_history (
			exchange_id,
			total_balance
		)
		VALUES ($1, $2)`,
		exchangeID,
		balance,
	)
	if err != nil {
		return fmt.Errorf("repository: add balance history: %w", err)
	}

	return nil
}

func (r *RepositoryExchanges) GetBalanceOneHourAgo(ctx context.Context, exchangeID string) (float64, bool, error) {
	if exchangeID == "" {
		return 0, false, fmt.Errorf("repository: exchange id is empty")
	}

	var balance float64

	err := r.db.QueryRow(ctx,
		`SELECT total_balance
		FROM exchange_balance_history
		WHERE exchange_id = $1
		  AND created_at <= NOW() - INTERVAL '1 hour'
		ORDER BY created_at DESC
		LIMIT 1`,
		exchangeID,
	).Scan(&balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("repository: get balance one hour ago: %w", err)
	}

	return balance, true, nil
}
