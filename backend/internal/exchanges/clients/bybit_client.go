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

type bybitResponse struct {
	RetCode int         `json:"retCode"`
	RetMsg  string      `json:"retMsg"`
	Result  bybitResult `json:"result"`
}

type bybitResult struct {
	List []bybitAccount `json:"list"`
}

type bybitAccount struct {
	TotalEquity string      `json:"totalEquity"`
	Coin        []bybitCoin `json:"coin"`
}

type bybitCoin struct {
	Coin          string `json:"coin"`
	Equity        string `json:"equity"`
	WalletBalance string `json:"walletBalance"`
	UsdValue      string `json:"usdValue"`
}

type BybitClient struct {
	httpClient *http.Client
}

func NewBybitClient(httpClient *http.Client) *BybitClient {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
		}
	}

	return &BybitClient{
		httpClient: httpClient,
	}
}

func (b *BybitClient) ValidateAndGetBalance(ctx context.Context, apiKey, apiSecret string) (ExchangeBalanceResult, error) {
	if apiKey == "" || apiSecret == "" {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	url := "https://api.bybit.com/v5/account/wallet-balance?accountType=UNIFIED"

	query := "accountType=UNIFIED"
	recvWindow := "20000"
	timestamp := time.Now().UnixMilli()
	signPayload := fmt.Sprintf("%d%s%s%s", timestamp, apiKey, recvWindow, query)

	hash := hmac.New(sha256.New, []byte(apiSecret))
	hash.Write([]byte(signPayload))
	signature := hex.EncodeToString(hash.Sum(nil))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error creating bybit request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-BAPI-SIGN", signature)
	req.Header.Set("X-BAPI-API-KEY", apiKey)
	req.Header.Set("X-BAPI-TIMESTAMP", strconv.FormatInt(timestamp, 10))
	req.Header.Set("X-BAPI-RECV-WINDOW", recvWindow)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("%w: %v", ErrExchangeUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ExchangeBalanceResult{}, ErrInvalidCredentials
	}

	if resp.StatusCode != http.StatusOK {
		return ExchangeBalanceResult{}, fmt.Errorf(
			"%w: bybit returned status %d",
			ErrExchangeUnavailable,
			resp.StatusCode,
		)
	}

	var res bybitResponse
	err = json.NewDecoder(resp.Body).Decode(&res)
	if err != nil {
		return ExchangeBalanceResult{}, fmt.Errorf("error decoding bybit response: %w", err)
	}

	if res.RetCode != 0 {
		return ExchangeBalanceResult{}, fmt.Errorf(
			"%w: bybit retCode=%d retMsg=%s",
			ErrInvalidCredentials,
			res.RetCode,
			res.RetMsg,
		)
	}

	if len(res.Result.List) == 0 {
		return ExchangeBalanceResult{}, fmt.Errorf("%w: empty bybit balance list", ErrExchangeUnavailable)
	}

	account := res.Result.List[0]

	sum := 0.0
	count := 0
	pairs := []Pair{}

	for _, coin := range account.Coin {
		asset := strings.ToUpper(strings.TrimSpace(coin.Coin))
		if asset == "" {
			continue
		}

		if coin.UsdValue == "" {
			continue
		}

		valueUSDT, err := strconv.ParseFloat(coin.UsdValue, 64)
		if err != nil {
			return ExchangeBalanceResult{}, fmt.Errorf("couldn't convert usdValue to float: %w", err)
		}

		if valueUSDT < minAssetValueUSDT {
			continue
		}

		amount := 0.0

		if coin.Equity != "" {
			amount, _ = strconv.ParseFloat(coin.Equity, 64)
		}

		if amount == 0 && coin.WalletBalance != "" {
			amount, _ = strconv.ParseFloat(coin.WalletBalance, 64)
		}

		priceUSDT := 0.0
		if amount > 0 {
			priceUSDT = valueUSDT / amount
		}

		count++

		pairs = append(pairs, Pair{
			Name:      asset,
			Amount:    amount,
			PriceUSDT: priceUSDT,
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
