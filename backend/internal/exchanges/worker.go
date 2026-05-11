package exchanges

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"cryptofund/internal/exchanges/clients"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

func updateDemoBalances() float64 {
	val := -100.0 + 200.0*rand.Float64()
	return math.Round(val*100) / 100
}

func calculatePercentageChange(newPrice, oldPrice float64) float64 {
	if oldPrice == 0 {
		if newPrice == 0 {
			return 0
		}

		if newPrice > 0 {
			return 100.0
		}

		return -100.0
	}

	percent := ((newPrice - oldPrice) / oldPrice) * 100
	percent = math.Round(percent*100) / 100

	return percent
}

func decipherKey(encryptedKey, masterKey string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("error decoding encrypted key: %w", err)
	}

	block, err := aes.NewCipher([]byte(masterKey))
	if err != nil {
		return "", fmt.Errorf("invalid master key: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creating gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted key is too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plainText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("error decrypting key: %w", err)
	}

	return string(plainText), nil
}

func (s *ServiceExchanges) UpdateWorker(ctx context.Context, userID string) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("balance worker stopped")
			return

		case <-ticker.C:
			err := s.RefreshUserBalances(ctx, userID)
			if err != nil {
				log.Printf("error refreshing user balances: %v", err)
			}
		}
	}
}

func (s *ServiceExchanges) RefreshUserBalances(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("userID is empty")
	}

	exchangesForUpdate, err := s.repo.GetExchangesForUpdateByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting exchanges for update: %w", err)
	}

	var wg sync.WaitGroup

	for _, ex := range exchangesForUpdate {
		ex := ex

		wg.Add(1)

		go func() {
			defer wg.Done()

			err := s.refreshOneExchange(ctx, ex)
			if err != nil {
				log.Printf(
					"error refreshing exchange id=%s name=%s source=%s: %v",
					ex.ID,
					ex.Name,
					ex.Source,
					err,
				)
			}
		}()
	}

	wg.Wait()

	return nil
}

func (s *ServiceExchanges) refreshOneExchange(ctx context.Context, ex ExchangeForUpdate) error {
	source := strings.ToLower(strings.TrimSpace(ex.Source))

	switch source {
	case "mock":
		return s.refreshMockExchange(ctx, ex)

	case "live":
		return s.refreshLiveExchange(ctx, ex)

	default:
		return fmt.Errorf("unsupported exchange source: %s", ex.Source)
	}
}

func (s *ServiceExchanges) refreshMockExchange(ctx context.Context, ex ExchangeForUpdate) error {
	newBalance := ex.Balance + updateDemoBalances()
	if newBalance < 0 {
		newBalance = 0
	}

	percentChange, err := s.calculateOneHourChangePercent(ctx, ex.ID, newBalance)
	if err != nil {
		return fmt.Errorf("error calculating mock one hour change: %w", err)
	}

	err = s.repo.AddBalanceByExchangeID(
		ctx,
		ex.ID,
		newBalance,
		percentChange,
		ex.AssetsCount,
		"mock",
	)
	if err != nil {
		return fmt.Errorf("error updating mock balance: %w", err)
	}

	err = s.repo.AddBalanceHistory(ctx, ex.ID, newBalance)
	if err != nil {
		return fmt.Errorf("error adding mock balance history: %w", err)
	}

	return nil
}

func (s *ServiceExchanges) refreshLiveExchange(ctx context.Context, ex ExchangeForUpdate) error {
	exchangeKey := strings.ToLower(strings.TrimSpace(ex.Name))

	cl, ok := s.client[exchangeKey]
	if !ok {
		return clients.ErrUnsupportedExchange
	}

	apiKey, err := decipherKey(ex.KeyAPI, s.masterKey)
	if err != nil {
		return fmt.Errorf("error decrypting api key: %w", err)
	}

	apiSecret, err := decipherKey(ex.SecretAPI, s.masterKey)
	if err != nil {
		return fmt.Errorf("error decrypting api secret: %w", err)
	}

	res, err := cl.ValidateAndGetBalance(ctx, apiKey, apiSecret)
	if err != nil {
		return fmt.Errorf("error getting live balance: %w", err)
	}

	source := res.Source
	if source == "" {
		source = "live"
	}

	percentChange, err := s.calculateOneHourChangePercent(ctx, ex.ID, res.Balance)
	if err != nil {
		return fmt.Errorf("error calculating live one hour change: %w", err)
	}

	err = s.repo.AddBalanceByExchangeID(
		ctx,
		ex.ID,
		res.Balance,
		percentChange,
		res.AssetsCount,
		source,
	)
	if err != nil {
		return fmt.Errorf("error updating live balance: %w", err)
	}

	err = s.repo.AddBalanceHistory(ctx, ex.ID, res.Balance)
	if err != nil {
		return fmt.Errorf("error adding live balance history: %w", err)
	}

	return nil
}

func (s *ServiceExchanges) calculateOneHourChangePercent(ctx context.Context, exchangeID string, currentBalance float64) (float64, error) {
	balanceOneHourAgo, found, err := s.repo.GetBalanceOneHourAgo(ctx, exchangeID)
	if err != nil {
		return 0, fmt.Errorf("error getting balance one hour ago: %w", err)
	}

	if !found {
		return 0, nil
	}

	return calculatePercentageChange(currentBalance, balanceOneHourAgo), nil
}
