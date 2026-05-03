package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockDB struct {
	execResult    pgconn.CommandTag
	execError     error
	queryRowValue string
	queryRowError error
}

func (m *mockDB) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	return m.execResult, m.execError
}

func (m *mockDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return &mockRow{val: m.queryRowValue, err: m.queryRowError}
}

type mockRow struct {
	val string
	err error
}

func (r *mockRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if ptr, ok := dest[0].(*string); ok {
		*ptr = r.val
	}
	return nil
}

func TestRepositoryAuth_AddUser_Success(t *testing.T) {
	mock := &mockDB{
		execResult: pgconn.NewCommandTag("INSERT 0 1"),
	}
	repo := NewRepository(mock)

	err := repo.AddUser(context.Background(), "test@test.com", "hash123")

	if err != nil {
		t.Fatalf("got error: %v", err)
	}
}

func TestRepositoryAuth_AddUser_Duplicate(t *testing.T) {
	mock := &mockDB{
		execResult: pgconn.NewCommandTag("INSERT 0 0"),
	}
	repo := NewRepository(mock)

	err := repo.AddUser(context.Background(), "test@test.com", "hash123")
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestRepositoryAuth_GetUserID_Success(t *testing.T) {
	mock := &mockDB{
		queryRowValue: "user-uuid-123",
	}
	repo := NewRepository(mock)

	id, err := repo.GetUserID(context.Background(), "test@test.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if id != "user-uuid-123" {
		t.Errorf("expected 'user-uuid-123', got: %s", id)
	}
}

func TestRepositoryAuth_GetUserID_NotFound(t *testing.T) {
	mock := &mockDB{
		queryRowError: pgx.ErrNoRows,
	}
	repo := NewRepository(mock)

	_, err := repo.GetUserID(context.Background(), "noone@test.com")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestRepositoryAuth_GetPasswordHash_Success(t *testing.T) {
	mock := &mockDB{
		queryRowValue: "$2a$10$testhash123",
	}
	repo := NewRepository(mock)

	hash, err := repo.GetPasswordHash(context.Background(), "test@test.com")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if hash != "$2a$10$testhash123" {
		t.Errorf("expected '$2a$10$testhash123', got: %s", hash)
	}
}

func TestRepositoryAuth_GetPasswordHash_Error(t *testing.T) {
	mock := &mockDB{
		queryRowError: errors.New("DB connection lost"),
	}
	repo := NewRepository(mock)

	_, err := repo.GetPasswordHash(context.Background(), "test@test.com")
	if err == nil {
		t.Error("expected error, got nil")
	}
}
