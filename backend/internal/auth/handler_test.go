package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockService struct {
	regResp ResponseAuth
	regErr  error
	logResp ResponseAuth
	logErr  error
}

func (m *mockService) RegisterService(ctx context.Context, data RegisterRequest) (ResponseAuth, error) {
	return m.regResp, m.regErr
}

func (m *mockService) LoginService(ctx context.Context, data LoginRequest) (ResponseAuth, error) {
	return m.logResp, m.logErr
}

func parseResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return result
}

func TestHandler_RegisterHandler_Success(t *testing.T) {
	mock := &mockService{
		regResp: ResponseAuth{Token: "test-token", UserID: "user-123", Email: "test@mail.ru"},
	}
	h := NewHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email":"test@mail.ru","password":"secure123"}`))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	resp := parseResponse(t, w)
	if resp["token"] != "test-token" {
		t.Errorf("expected token 'test-token', got %v", resp["token"])
	}
}

func TestHandler_RegisterHandler_ValidationErrors(t *testing.T) {
	h := NewHandler(&mockService{})

	tests := []struct {
		name   string
		body   string
		expErr string
	}{
		{"invalid_json", `{bad json}`, "invalid_json"},
		{"empty_email", `{"email":"","password":"123"}`, "empty_email"},
		{"empty_password", `{"email":"test@mail.ru","password":""}`, "empty_password"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			h.RegisterHandler(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: expected status %d, got %d", tt.name, http.StatusBadRequest, w.Code)
			}

			resp := parseResponse(t, w)
			if resp["error"] != tt.expErr {
				t.Errorf("%s: expected error '%s', got '%s'", tt.name, tt.expErr, resp["error"])
			}
		})
	}
}

func TestHandler_RegisterHandler_ServiceError(t *testing.T) {
	mock := &mockService{regErr: ErrUserAlreadyExists}
	h := NewHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email":"exist@mail.ru","password":"123"}`))
	w := httptest.NewRecorder()

	h.RegisterHandler(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	resp := parseResponse(t, w)
	if resp["error"] != "user_exists" {
		t.Errorf("expected error 'user_exists', got '%s'", resp["error"])
	}
}

func TestHandler_LoginHandler_Success(t *testing.T) {
	mock := &mockService{
		logResp: ResponseAuth{Token: "login-token", UserID: "user-456", Email: "login@mail.ru"},
	}
	h := NewHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"login@mail.ru","password":"secure123"}`))
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	resp := parseResponse(t, w)
	if resp["token"] != "login-token" {
		t.Errorf("expected token 'login-token', got %v", resp["token"])
	}
}

func TestHandler_LoginHandler_ServiceError(t *testing.T) {
	mock := &mockService{logErr: ErrInvalidCredentials}
	h := NewHandler(mock)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"nope@mail.ru","password":"wrong"}`))
	w := httptest.NewRecorder()

	h.LoginHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	resp := parseResponse(t, w)
	if resp["error"] != "invalid_credentials" {
		t.Errorf("expected error 'invalid_credentials', got '%s'", resp["error"])
	}
}
