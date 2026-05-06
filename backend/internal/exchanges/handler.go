package exchanges

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type Service interface {
	AddExchange(ctx context.Context, userID, name, keyAPI, secretAPI string) (ExchangeCreateResponse, error)
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
		log.Printf("AddExchange failed: user_id=%s name=%s err=%v", userID, exchangeData.Name, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal_server_error"})
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}
