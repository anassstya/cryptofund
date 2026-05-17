package clients

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const minAssetValueUSDT = 0.5

type mexcAccountResponse struct {
	Balances []mexcBalance `json:"balances"`
}

type mexcBalance struct {
	Asset     string `json:"asset"`
	Free      string `json:"free"`
	Locked    string `json:"locked"`
	Available string `json:"available"`
}

type PriceData struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type PriceCache interface {
	GetPrice(ctx context.Context, key string) (float64, error)
	SetPrice(ctx context.Context, key string, value float64) error
}

type MexcClient struct {
	httpClient *http.Client
	priceCache PriceCache
}

func NewMexcClient(httpClient *http.Client, priceCache ...PriceCache) *MexcClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	var cache PriceCache
	if len(priceCache) > 0 {
		cache = priceCache[0]
	}

	return &MexcClient{
		httpClient: httpClient,
		priceCache: cache,
	}
}

func priceCacheKey(exchangeName, asset string) string {
	return fmt.Sprintf(
		"price:%s:%s",
		strings.ToLower(exchangeName),
		strings.ToUpper(asset),
	)
}

func (m *MexcClient) ValidateAndGetBalance(ctx context.Context, apiKey, apiSecret string) (ExchangeBalanceResult, error) {
	if apiKey == "" || apiSecret == "" {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d", timestamp)

	hash := hmac.New(sha256.New, []byte(apiSecret))
	hash.Write([]byte(query))
	signature := hex.EncodeToString(hash.Sum(nil))

	baseURL := "https://api.mexc.com"
	url := fmt.Sprintf("%s/api/v3/account?%s&signature=%s", baseURL, query, signature)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error creating mexc request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MEXC-APIKEY", apiKey)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	if resp.StatusCode != http.StatusOK {
		return ExchangeBalanceResult{}, fmt.Errorf(
			"%w: mexc returned status %d",
			ErrExchangeUnavailable,
			resp.StatusCode,
		)
	}

	var data mexcAccountResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error decoding mexc response: %w", err)
	}

	sum := 0.0
	count := 0
	pairs := []Pair{}

	for _, v := range data.Balances {
		asset := strings.ToUpper(strings.TrimSpace(v.Asset))
		if asset == "" {
			continue
		}

		free, err := strconv.ParseFloat(v.Free, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert free balance to float: %w", err)
		}

		locked, err := strconv.ParseFloat(v.Locked, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert locked balance to float: %w", err)
		}

		amount := free + locked
		if amount <= 0 {
			continue
		}

		price := 0.0

		if asset == "USDT" {
			price = 1
			log.Printf("MEXC price %s loaded directly as stablecoin: %.8f", asset, price)
		} else if m.priceCache != nil {
			key := priceCacheKey("mexc", asset)

			cachedPrice, err := m.priceCache.GetPrice(ctx, key)
			if err == nil {
				price = cachedPrice
			}
		}

		if price == 0 {
			price, err = m.GetAssetValueUSDT(ctx, asset)
			if err != nil {
				log.Printf("MEXC price %s API loading failed: %v", asset, err)
				continue
			}

			if m.priceCache != nil {
				key := priceCacheKey("mexc", asset)

				if err := m.priceCache.SetPrice(ctx, key, price); err != nil {
					log.Printf("MEXC price %s failed to save to Redis cache: %v", asset, err)
				}
			}
		}

		valueUSDT := amount * price

		if valueUSDT < minAssetValueUSDT {
			continue
		}

		count++

		pairs = append(pairs, Pair{
			Name:      asset,
			Amount:    amount,
			PriceUSDT: price,
			ValueUSDT: valueUSDT,
		})

		sum += valueUSDT
	}

	return ExchangeBalanceResult{
		Balance:     sum,
		AssetsCount: count,
		Source:      "live",
		Pairs:       pairs,
	}, nil
}

func (m *MexcClient) GetAssetValueUSDT(ctx context.Context, name string) (float64, error) {
	asset := strings.ToUpper(strings.TrimSpace(name))

	if asset == "" {
		return 0, fmt.Errorf("asset name is empty")
	}

	if asset == "USDT" {
		return 1, nil
	}

	url := fmt.Sprintf("https://api.mexc.com/api/v3/ticker/price?symbol=%sUSDT", asset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("couldn't create mexc price request: %w", err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("couldn't find pair %sUSDT", asset)
	}

	var res PriceData

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, fmt.Errorf("couldn't decode price: %w", err)
	}

	p, err := strconv.ParseFloat(res.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't convert price to float: %w", err)
	}

	return p, nil
}
