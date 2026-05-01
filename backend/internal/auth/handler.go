package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method_not_allowed"})
		return
	}

	var registerData RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&registerData)

	if err != nil {
		log.Printf("error decoding: %v", err)
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]string{
			"error":   "invalid_json",
			"message": "Не удалось распарсить запрос. Проверьте формат JSON.",
		})
		return
	}

	if registerData.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "empty_email",
			"message": "Email is required",
		})
		return
	}

	if registerData.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "empty_password",
			"message": "Password is required",
		})
		return
	}

	resp, err := h.service.RegisterService(r.Context(), registerData)

	if err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error":   "user_exists",
				"message": "Пользователь с таким email уже зарегистрирован",
			})
			return
		}

		log.Printf("register error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "internal_error"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {

	}
}
