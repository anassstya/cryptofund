package clients

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type gateAccountBalance struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

type gateTickerResponse struct {
	CurrencyPair string `json:"currency_pair"`
	Last         string `json:"last"`
}

type gateErrorResponse struct {
	Label   string `json:"label"`
	Message string `json:"message"`
}

type GateClient struct {
	httpClient *http.Client
	priceCache PriceCache
}

func NewGateClient(httpClient *http.Client, priceCache ...PriceCache) *GateClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	var cache PriceCache
	if len(priceCache) > 0 {
		cache = priceCache[0]
	}

	return &GateClient{
		httpClient: httpClient,
		priceCache: cache,
	}
}

func (g *GateClient) ValidateAndGetBalance(ctx context.Context, apiKey, apiSecret string) (ExchangeBalanceResult, error) {
	if apiKey == "" || apiSecret == "" {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	method := http.MethodGet
	requestPath := "/spot/accounts"
	query := ""
	body := ""

	url := "https://api.gateio.ws/api/v4" + requestPath

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := signGate(method, requestPath, query, body, timestamp, apiSecret)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error creating gate request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("KEY", apiKey)
	req.Header.Set("Timestamp", timestamp)
	req.Header.Set("SIGN", signature)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	if resp.StatusCode != http.StatusOK {
		var gateErr gateErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&gateErr); err == nil && gateErr.Message != "" {
			return ExchangeBalanceResult{}, fmt.Errorf(
				"%w: gate label=%s message=%s",
				ErrExchangeUnavailable,
				gateErr.Label,
				gateErr.Message,
			)
		}

		return ExchangeBalanceResult{}, fmt.Errorf(
			"%w: gate returned status %d",
			ErrExchangeUnavailable,
			resp.StatusCode,
		)
	}

	var balances []gateAccountBalance

	if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error decoding gate response: %w", err)
	}

	sum := 0.0
	count := 0
	pairs := []Pair{}

	for _, balance := range balances {
		asset := strings.ToUpper(strings.TrimSpace(balance.Currency))
		if asset == "" {
			continue
		}

		available, err := strconv.ParseFloat(balance.Available, 64)
		if err != nil {
			continue
		}

		locked, err := strconv.ParseFloat(balance.Locked, 64)
		if err != nil {
			continue
		}

		amount := available + locked
		if amount <= 0 {
			continue
		}

		price, err := g.GetAssetPriceUSDT(ctx, asset)
		if err != nil {
			continue
		}

		valueUSDT := price * amount

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
		Balance:       sum,
		ChangePercent: 0,
		AssetsCount:   count,
		Source:        "live",
		Pairs:         pairs,
	}, nil
}

func (g *GateClient) GetAssetPriceUSDT(ctx context.Context, currency string) (float64, error) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		return 0, fmt.Errorf("currency is empty")
	}

	if currency == "USDT" {
		return 1, nil
	}

	if g.priceCache != nil {
		key := priceCacheKey("gate", currency)

		cachedPrice, err := g.priceCache.GetPrice(ctx, key)
		if err == nil {
			return cachedPrice, nil
		}
	}

	currencyPair := currency + "_USDT"

	url := fmt.Sprintf(
		"https://api.gateio.ws/api/v4/spot/tickers?currency_pair=%s",
		currencyPair,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating gate price request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("couldn't find gate pair %s", currencyPair)
	}

	var tickers []gateTickerResponse

	if err := json.NewDecoder(resp.Body).Decode(&tickers); err != nil {
		return 0, fmt.Errorf("couldn't decode gate price: %w", err)
	}

	if len(tickers) == 0 {
		return 0, fmt.Errorf("couldn't find gate pair %s", currencyPair)
	}

	price, err := strconv.ParseFloat(tickers[0].Last, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't convert gate price to float: %w", err)
	}

	if g.priceCache != nil {
		key := priceCacheKey("gate", currency)
		_ = g.priceCache.SetPrice(ctx, key, price)
	}

	return price, nil
}

func signGate(method, requestPath, query, body, timestamp, apiSecret string) string {
	hashedBody := sha512.Sum512([]byte(body))
	hashedBodyHex := hex.EncodeToString(hashedBody[:])

	signatureString := method + "\n" +
		requestPath + "\n" +
		query + "\n" +
		hashedBodyHex + "\n" +
		timestamp

	mac := hmac.New(sha512.New, []byte(apiSecret))
	mac.Write([]byte(signatureString))

	return hex.EncodeToString(mac.Sum(nil))
}
