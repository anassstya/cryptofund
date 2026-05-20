package exchanges

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"cryptofund/internal/exchanges/clients"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type Repository interface {
	AddExchange(ctx context.Context, userID, name, keyAPI, secretAPI, credentialsHash string) (string, error)
	GetByUserID(ctx context.Context, userID string) ([]ExchangeCreateResponse, error)
	AddBalanceByExchangeID(ctx context.Context, id string, balance float64, changePercent float64, assetsCount int, source string, pairs []clients.Pair) error

	GetBalanceByExchangeID(ctx context.Context, id string) (ExchangeBalanceResponse, error)

	GetBalanceOneHourAgo(ctx context.Context, exchangeID string) (float64, bool, error)
	AddBalanceHistory(ctx context.Context, exchangeID string, balance float64) error

	GetExchangesForUpdateByUserID(ctx context.Context, userID string) ([]ExchangeForUpdate, error)
}

type ServiceExchanges struct {
	repo      Repository
	masterKey string
	client    map[string]clients.ExchangeClient
}

func NewService(repo Repository, masterKey string, priceCache ...clients.PriceCache) *ServiceExchanges {
	var cache clients.PriceCache
	if len(priceCache) > 0 {
		cache = priceCache[0]
	}

	return &ServiceExchanges{
		repo:      repo,
		masterKey: masterKey,
		client: map[string]clients.ExchangeClient{
			"mexc":    clients.NewMexcClient(nil, cache),
			"bybit":   clients.NewBybitClient(nil),
			"binance": clients.NewBinanceClient(nil, cache),
			"gate":    clients.NewGateClient(nil, cache),
		},
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

func findDemoExchange(name, apiKey, apiSecret string) (DemoExchange, bool) {
	data, err := os.ReadFile("tempData.json")
	if err != nil {
		fmt.Printf("error reading json: %v\n", err)
		return DemoExchange{}, false
	}

	var demos []DemoExchange

	err = json.Unmarshal(data, &demos)
	if err != nil {
		fmt.Printf("error unmarshaling json: %v\n", err)
		return DemoExchange{}, false
	}

	for _, v := range demos {
		if v.Name == name && v.APIKey == apiKey && v.APISecret == apiSecret {
			return v, true
		}
	}

	return DemoExchange{}, false
}

func makeCredentialsHash(name, apiKey, apiSecret string) string {
	sum := sha256.Sum256([]byte(name + ":" + apiKey + ":" + apiSecret))
	return hex.EncodeToString(sum[:])
}

func (s *ServiceExchanges) AddExchange(ctx context.Context, userID, name, keyAPI, secretAPI string) (ExchangeCreateResponse, error) {
	var balance clients.ExchangeBalanceResult

	data, isMock := findDemoExchange(name, keyAPI, secretAPI)

	if isMock {
		source := data.Source
		if source == "" {
			source = "mock"
		}

		realBalance := calculatePairsTotalUSDT(data.Pairs)

		balance = clients.ExchangeBalanceResult{
			Balance:       realBalance,
			ChangePercent: data.ChangePercent,
			AssetsCount:   len(data.Pairs),
			Source:        source,
			Pairs:         data.Pairs,
		}
	} else {
		exchangeKey := strings.ToLower(strings.TrimSpace(name))

		cl, ok := s.client[exchangeKey]
		if !ok {
			return ExchangeCreateResponse{}, clients.ErrUnsupportedExchange
		}

		res, err := cl.ValidateAndGetBalance(ctx, keyAPI, secretAPI)
		if err != nil {
			return ExchangeCreateResponse{}, fmt.Errorf("%w: %v", ErrInvalidCredentials, err)
		}

		source := res.Source
		if source == "" {
			source = "live"
		}

		balance = clients.ExchangeBalanceResult{
			Balance:       res.Balance,
			ChangePercent: res.ChangePercent,
			AssetsCount:   res.AssetsCount,
			Source:        source,
			Pairs:         res.Pairs,
		}
	}

	credentialsHash := makeCredentialsHash(name, keyAPI, secretAPI)

	key, err := cipherKey(keyAPI, s.masterKey)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error encrypting key: %w", err)
	}

	secret, err := cipherKey(secretAPI, s.masterKey)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error encrypting secret: %w", err)
	}

	id, err := s.repo.AddExchange(ctx, userID, name, key, secret, credentialsHash)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error adding exchange to db: %w", err)
	}

	err = s.repo.AddBalanceByExchangeID(
		ctx,
		id,
		balance.Balance,
		balance.ChangePercent,
		balance.AssetsCount,
		balance.Source,
		balance.Pairs,
	)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error adding balance to db: %w", err)
	}

	err = s.repo.AddBalanceHistory(ctx, id, balance.Balance)
	if err != nil {
		return ExchangeCreateResponse{}, fmt.Errorf("error adding balance history: %w", err)
	}

	return ExchangeCreateResponse{
		ID:    id,
		Name:  name,
		Pairs: balance.Pairs,
	}, nil
}

func (s *ServiceExchanges) GetByUserID(ctx context.Context, userID string) ([]ExchangeCreateResponse, error) {
	if userID == "" {
		return []ExchangeCreateResponse{}, fmt.Errorf("userID is empty")
	}

	res, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return []ExchangeCreateResponse{}, fmt.Errorf("error getting exchanges: %w", err)
	}

	return res, nil
}

func (s *ServiceExchanges) GetBalanceByExchangeID(ctx context.Context, exchangeID string) (ExchangeBalanceResponse, error) {
	if exchangeID == "" {
		return ExchangeBalanceResponse{}, fmt.Errorf("exchangeID is empty")
	}

	res, err := s.repo.GetBalanceByExchangeID(ctx, exchangeID)
	if err != nil {
		return ExchangeBalanceResponse{}, fmt.Errorf("error getting balance: %w", err)
	}

	return res, nil
}
