package exchanges

import (
	"cryptofund/internal/exchanges/clients"
	"time"
)

type Exchange struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	KeyAPI    string `json:"api_key"`
	SecretAPI string `json:"api_secret"`
}

type ExchangeCreateResponse struct {
	ID    string         `json:"id"`
	Name  string         `json:"name"`
	Pairs []clients.Pair `json:"pairs,omitempty"`
}

type ExchangeBalanceResponse struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Balance       float64        `json:"total_balance"`
	ChangePercent float64        `json:"change_percent"`
	AssetsCount   int            `json:"assets_count"`
	Source        string         `json:"source"`
	UpdatedAt     time.Time      `json:"updated_at"`
	Pairs         []clients.Pair `json:"pairs"`
}

type DemoExchange struct {
	Name          string         `json:"name"`
	APIKey        string         `json:"api_key"`
	APISecret     string         `json:"api_secret"`
	Balance       float64        `json:"total_balance"`
	ChangePercent float64        `json:"change_percent"`
	AssetsCount   int            `json:"assets_count"`
	Source        string         `json:"source"`

	Pairs         []clients.Pair `json:"pairs"`
}

type ExchangeForUpdate struct {
	ID          string
	Name        string
	KeyAPI      string
	SecretAPI   string
	Balance     float64
	AssetsCount int
	Source      string
}
