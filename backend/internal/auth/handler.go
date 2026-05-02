package auth

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Service interface {
	RegisterService(ctx context.Context, data RegisterRequest) (ResponseAuth, error)
	LoginService(ctx context.Context, data LoginRequest) (ResponseAuth, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func writeAuthError(w http.ResponseWriter, err error, logPrefix string) {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		writeJSON(w, http.StatusConflict, map[string]string{
			"error":   "user_exists",
			"message": "Пользователь с таким email уже зарегистрирован",
		})
	case errors.Is(err, ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "invalid_credentials",
			"message": "Неверный email или пароль",
		})
	default:
		log.Printf("%s: %v", logPrefix, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_error"})
	}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var registerData RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerData)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_json",
			"message": "Couldn't parse JSON",
		})
		return
	}

	if registerData.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "empty_email",
			"message": "Email is required",
		})
		return
	}

	if registerData.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "empty_password",
			"message": "Password is required",
		})
		return
	}

	resp, err := h.service.RegisterService(r.Context(), registerData)
	if err != nil {
		writeAuthError(w, err, "register error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var loginData LoginRequest
	err := json.NewDecoder(r.Body).Decode(&loginData)

	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_json",
			"message": "Couldn't parse JSON",
		})
		return
	}

	if loginData.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "empty_email",
			"message": "Email is required",
		})
		return
	}

	if loginData.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "empty_password",
			"message": "Password is required",
		})
		return
	}

	resp, err := h.service.LoginService(r.Context(), loginData)
	if err != nil {
		writeAuthError(w, err, "login error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
