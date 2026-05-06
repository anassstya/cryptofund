package exchanges

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Repository interface {
	AddExchange(ctx context.Context, userID, name, KeyAPI, SecretAPI string) (string, error)
	GetByUserID(ctx context.Context, userID string) ([]Exchange, error)
}

type ServiceExchanges struct {
	repo      Repository
	masterKey string
}

func NewService(repo Repository, masterKey string) *ServiceExchanges {
	return &ServiceExchanges{
		repo:      repo,
		masterKey: masterKey,
	}
}

func cipherKey(key, masterKey string) (string, error) {
	block, err := aes.NewCipher([]byte(masterKey))
	if err != nil {
		return "", fmt.Errorf("invalid master key: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(key), nil)
	encoded := base64.StdEncoding.EncodeToString(append(nonce, ciphertext...))

	return encoded, nil
}

func (s *ServiceExchanges) AddExchange(ctx context.Context, userID, name, keyAPI, secretAPI string) (ExchangeCreateResponse, error) {
	key, err := cipherKey(keyAPI, s.masterKey)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error getting cipherted key: %w", err)
	}

	secret, err := cipherKey(secretAPI, s.masterKey)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error getting cipherted secret: %w", err)
	}

	id, err := s.repo.AddExchange(ctx, userID, name, key, secret)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error adding exchange to db: %w", err)
	}

	return ExchangeCreateResponse{
		ID:   id,
		Name: name,
	}, nil
}
