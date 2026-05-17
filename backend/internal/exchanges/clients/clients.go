package clients

import (
	"context"
	"errors"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrUnsupportedExchange = errors.New("unsupported exchange")
var ErrExchangeUnavailable = errors.New("exchange unavailable")

type ExchangeBalanceResult struct {
	Balance       float64
	ChangePercent float64
	AssetsCount   int
	Source        string

	Pairs []Pair
}

type Pair struct {
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	PriceUSDT float64 `json:"price_usdt"`
	ValueUSDT float64 `json:"value_usdt"`
}

type ExchangeClient interface {
	ValidateAndGetBalance(ctx context.Context, apiKey, apiSecret string) (ExchangeBalanceResult, error)
}
