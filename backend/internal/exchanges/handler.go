package exchanges

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Service interface {
	AddExchange(ctx context.Context, userID, name, keyAPI, secretAPI string) (ExchangeCreateResponse, error)
	GetByUserID(ctx context.Context, userID string) ([]ExchangeCreateResponse, error)
	GetBalanceByExchangeID(ctx context.Context, exchangeID string) (ExchangeBalanceResponse, error)

	RefreshUserBalances(ctx context.Context, userID string) error
}

type HandlerExchanges struct {
	serv Service
}

func NewHandler(serv Service) *HandlerExchanges {
	return &HandlerExchanges{
		serv: serv,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode JSON response: %v", err)
	}
}

func (s *HandlerExchanges) AddExchangeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	var exchangeData Exchange
	err := json.NewDecoder(r.Body).Decode(&exchangeData)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "invalid_json",
			"message": "Couldn't parse JSON"})
		return
	}

	if exchangeData.Name == "" || exchangeData.KeyAPI == "" || exchangeData.SecretAPI == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing_required_fields"})
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	resp, err := s.serv.AddExchange(r.Context(), userID, exchangeData.Name, exchangeData.KeyAPI, exchangeData.SecretAPI)

	if err != nil {
		if errors.Is(err, ErrExchangeAlreadyExists) {
			log.Printf("ошибка в добавлении в ьд: %w", ErrExchangeAlreadyExists)
			writeJSON(w, http.StatusConflict, map[string]string{
				"error":   "exchange_already_exists",
				"message": "Такая биржа уже добавлена",
			})
			return
		}

		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (s *HandlerExchanges) GetExchangesWithBalanceHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method_not_allowed"})
		return
	}

	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	err := s.serv.RefreshUserBalances(r.Context(), userID)
	if err != nil {
		log.Printf("refresh user balances failed: user_id=%s err=%v", userID, err)
	}

	usersExchanges, err := s.serv.GetByUserID(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error getting exchanges"})
		return
	}

	res := make([]ExchangeBalanceResponse, 0)

	for _, v := range usersExchanges {
		log.Printf("🔎 Обработка биржи: v.ID='%s', v.Name='%s'", v.ID, v.Name)
		a, err := s.serv.GetBalanceByExchangeID(r.Context(), v.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error getting balance"})
			return
		}

		res = append(res, a)
	}

	writeJSON(w, http.StatusOK, res)
}
