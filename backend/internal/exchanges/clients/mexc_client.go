package clients

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

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

type MexcClient struct {
	httpClient *http.Client
}

func NewMexcClient(httpClient *http.Client) *MexcClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &MexcClient{
		httpClient: httpClient,
	}
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
		return ExchangeBalanceResult{}, fmt.Errorf("error decoding %v", err)
	}

	balances := data.Balances

	sum := 0.0
	count := 0
	for _, v := range balances {
		free, err := strconv.ParseFloat(v.Free, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert to float %v", err)
		}

		locked, err := strconv.ParseFloat(v.Locked, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert to float %v", err)
		}

		if free+locked > 0 {
			count++

			p, err := m.GetAssetValueUSDT(ctx, v.Asset, free, locked)
			if err != nil {
				continue
			}
			sum += p
		}
	}

	return ExchangeBalanceResult{
		Balance:     sum,
		AssetsCount: count,
		Source:      "live",
	}, nil
}

func (m *MexcClient) GetAssetValueUSDT(ctx context.Context, name string, free, locked float64) (float64, error) {
	if name == "USDT" {
		return free + locked, nil
	}
	url := fmt.Sprintf("https://api.mexc.com/api/v3/ticker/price?symbol=%vUSDT", name)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("couldn't get price")
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("couldn't find pair")
	}

	var res PriceData

	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return 0, fmt.Errorf("couldn't decode price")
	}

	p, err := strconv.ParseFloat(res.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't convert to float")
	}

	return p * (free + locked), nil
}
