package exchanges

type Exchange struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	KeyAPI    string `json:"api_key"`
	SecretAPI string `json:"api_secret"`
}

type ExchangeCreateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ExchangeResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Balance       float64 `json:"balance"`
	ChangePercent float64 `json:"change_percent"`
	AssetsCount   int     `json:"assets_count"`
}
