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
}

type ExchangeClient interface {
	ValidateAndGetBalance(ctx context.Context, apiKey, apiSecret string) (ExchangeBalanceResult, error)
}
