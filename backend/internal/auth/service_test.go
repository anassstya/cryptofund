package auth

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	addUserErr            error
	getUserIDResult       string
	getUserIDErr          error
	getPasswordHashResult string
	getPasswordHashErr    error
}

func (m *mockRepository) AddUser(ctx context.Context, email, hash string) error {
	return m.addUserErr
}

func (m *mockRepository) GetUserID(ctx context.Context, email string) (string, error) {
	return m.getUserIDResult, m.getUserIDErr
}

func (m *mockRepository) GetPasswordHash(ctx context.Context, email string) (string, error) {
	return m.getPasswordHashResult, m.getPasswordHashErr
}

func TestServiceAuth_RegisterService_Success(t *testing.T) {
	repo := &mockRepository{
		addUserErr:      nil,
		getUserIDResult: "user-uuid-123",
	}
	svc := NewService(repo, "test-jwt-secret-key-must-be-at-least-32-chars-long!!")

	resp, err := svc.RegisterService(context.Background(), RegisterRequest{
		Email:    "newuser@test.com",
		Password: "SecurePass123!",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token to be generated")
	}
	if resp.UserID != "user-uuid-123" {
		t.Errorf("expected user ID 'user-uuid-123', got: %s", resp.UserID)
	}
	if resp.Email != "newuser@test.com" {
		t.Errorf("expected email to match, got: %s", resp.Email)
	}
}

func TestServiceAuth_RegisterService_UserAlreadyExists(t *testing.T) {
	repo := &mockRepository{
		addUserErr: ErrUserAlreadyExists,
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.RegisterService(context.Background(), RegisterRequest{
		Email:    "exist@test.com",
		Password: "123456",
	})

	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got: %v", err)
	}
}

func TestServiceAuth_RegisterService_DBError(t *testing.T) {
	repo := &mockRepository{
		addUserErr: errors.New("db connection failed"),
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.RegisterService(context.Background(), RegisterRequest{
		Email:    "fail@test.com",
		Password: "123456",
	})

	if err == nil {
		t.Error("expected error on DB failure")
	}
	if errors.Is(err, ErrUserAlreadyExists) {
		t.Error("error should not be ErrUserAlreadyExists")
	}
}

func TestServiceAuth_LoginService_Success(t *testing.T) {
	password := "CorrectPassword!"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	repo := &mockRepository{
		getPasswordHashResult: string(hash),
		getUserIDResult:       "user-456",
	}
	svc := NewService(repo, "test-jwt-secret-key-must-be-at-least-32-chars-long!!")

	resp, err := svc.LoginService(context.Background(), LoginRequest{
		Email:    "user@test.com",
		Password: password,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if resp.Token == "" {
		t.Error("expected token")
	}
	if resp.UserID != "user-456" {
		t.Errorf("expected user ID 'user-456', got: %s", resp.UserID)
	}
}

func TestServiceAuth_LoginService_WrongPassword(t *testing.T) {
	password := "CorrectPassword!"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	repo := &mockRepository{
		getPasswordHashResult: string(hash),
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.LoginService(context.Background(), LoginRequest{
		Email:    "user@test.com",
		Password: "WrongPassword",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestServiceAuth_LoginService_UserNotFound(t *testing.T) {
	repo := &mockRepository{
		getPasswordHashErr: ErrUserNotFound,
	}
	svc := NewService(repo, "test-secret")

	_, err := svc.LoginService(context.Background(), LoginRequest{
		Email:    "noone@test.com",
		Password: "anypass",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for not found user, got: %v", err)
	}
}

func TestServiceAuth_generateToken_Valid(t *testing.T) {
	svc := NewService(&mockRepository{}, "test-jwt-secret-key-must-be-at-least-32-chars-long!!")

	token, err := svc.generateToken("uid-1", "test@mail.ru")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}
