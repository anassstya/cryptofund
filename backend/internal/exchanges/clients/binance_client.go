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
	"strings"
	"time"
)

type binanceAccountResponse struct {
	Balances []binanceBalance `json:"balances"`
}

type binanceBalance struct {
	Asset  string `json:"asset"`
	Free   string `json:"free"`
	Locked string `json:"locked"`
}

type binancePriceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

type binanceErrorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type BinanceClient struct {
	httpClient *http.Client
	priceCache PriceCache
}

func NewBinanceClient(httpClient *http.Client, priceCache ...PriceCache) *BinanceClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	var cache PriceCache
	if len(priceCache) > 0 {
		cache = priceCache[0]
	}

	return &BinanceClient{
		httpClient: httpClient,
		priceCache: cache,
	}
}

func (b *BinanceClient) ValidateAndGetBalance(ctx context.Context, apiKey, apiSecret string) (ExchangeBalanceResult, error) {
	if apiKey == "" || apiSecret == "" {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	timestamp := time.Now().UnixMilli()
	query := fmt.Sprintf("timestamp=%d", timestamp)

	hash := hmac.New(sha256.New, []byte(apiSecret))
	hash.Write([]byte(query))
	signature := hex.EncodeToString(hash.Sum(nil))

	url := fmt.Sprintf(
		"https://api.binance.com/api/v3/account?%s&signature=%s",
		query,
		signature,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error creating binance request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-MBX-APIKEY", apiKey)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	if resp.StatusCode != http.StatusOK {
		var binanceErr binanceErrorResponse

		if err := json.NewDecoder(resp.Body).Decode(&binanceErr); err == nil && binanceErr.Msg != "" {
			return ExchangeBalanceResult{}, fmt.Errorf(
				"%w: binance code=%d msg=%s",
				ErrInvalidCredentials,
				binanceErr.Code,
				binanceErr.Msg,
			)
		}

		return ExchangeBalanceResult{}, fmt.Errorf(
			"%w: binance returned status %d",
			ErrExchangeUnavailable,
			resp.StatusCode,
		)
	}

	var data binanceAccountResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error decoding binance response: %w", err)
	}

	sum := 0.0
	count := 0

	for _, balance := range data.Balances {
		free, err := strconv.ParseFloat(balance.Free, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert free balance to float: %w", err)
		}

		locked, err := strconv.ParseFloat(balance.Locked, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert locked balance to float: %w", err)
		}

		amount := free + locked
		if amount <= 0 {
			continue
		}

		count++

		price, err := b.GetAssetPriceUSDT(ctx, balance.Asset)
		if err != nil {
			continue
		}

		sum += price * amount
	}

	return ExchangeBalanceResult{
		Balance:       sum,
		ChangePercent: 0,
		AssetsCount:   count,
		Source:        "live",
	}, nil
}

func (b *BinanceClient) GetAssetPriceUSDT(ctx context.Context, asset string) (float64, error) {
	asset = strings.ToUpper(strings.TrimSpace(asset))
	if asset == "" {
		return 0, fmt.Errorf("asset is empty")
	}

	if asset == "USDT" {
		return 1, nil
	}

	if b.priceCache != nil {
		key := priceCacheKey("binance", asset)

		cachedPrice, err := b.priceCache.GetPrice(ctx, key)
		if err == nil {
			return cachedPrice, nil
		}
	}

	url := fmt.Sprintf(
		"https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT",
		asset,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating binance price request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("couldn't find binance pair %sUSDT", asset)
	}

	var priceData binancePriceResponse

	err = json.NewDecoder(resp.Body).Decode(&priceData)
	if err != nil {
		return 0, fmt.Errorf("couldn't decode binance price: %w", err)
	}

	price, err := strconv.ParseFloat(priceData.Price, 64)
	if err != nil {
		return 0, fmt.Errorf("couldn't convert binance price to float: %w", err)
	}

	if b.priceCache != nil {
		key := priceCacheKey("binance", asset)
		_ = b.priceCache.SetPrice(ctx, key, price)
	}

	return price, nil
}
