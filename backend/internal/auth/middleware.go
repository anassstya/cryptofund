package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func MiddlewareAuth(h http.HandlerFunc) http.HandlerFunc {
	key := os.Getenv("JWT_SECRET")

	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		edited, ok := strings.CutPrefix(token, "Bearer ")
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		t, err := jwt.Parse(edited, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(key), nil
		})

		if err != nil || !t.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok || userID == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", userID)
		h(w, r.WithContext(ctx))
	}
}
