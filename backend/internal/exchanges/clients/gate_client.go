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
}

func NewGateClient(httpClient *http.Client) *GateClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &GateClient{
		httpClient: httpClient,
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

	for _, balance := range balances {
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

		count++

		value, err := g.GetAssetValueUSDT(ctx, balance.Currency, amount)
		if err != nil {
			continue
		}

		sum += value
	}

	return ExchangeBalanceResult{
		Balance:       sum,
		ChangePercent: 0,
		AssetsCount:   count,
		Source:        "live",
	}, nil
}

func (g *GateClient) GetAssetValueUSDT(ctx context.Context, currency string, amount float64) (float64, error) {
	if amount <= 0 {
		return 0, nil
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))

	if currency == "USDT" {
		return amount, nil
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

	return price * amount, nil
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
