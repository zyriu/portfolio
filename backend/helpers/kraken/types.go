package kraken

type Kraken struct {
	ApiKey    string
	ApiSecret string
}

type Balances struct {
	Error  []string          `json:"error"`
	Result map[string]string `json:"result"`
}

type TradesHistory struct {
	Error  []string `json:"error"`
	Result struct {
		Count  int64 `json:"count"`
		Trades map[string]struct {
			TradeID   int64   `json:"trade_id"`
			Pair      string  `json:"pair"`
			Time      float64 `json:"time"`
			Type      string  `json:"type"`
			OrderType string  `json:"ordertype"`
			Price     string  `json:"price"`
			Cost      string  `json:"cost"`
			Fee       string  `json:"fee"`
			Vol       string  `json:"vol"`
			Net       string  `json:"net"`
			Maker     bool    `json:"maker"`
		} `json:"trades"`
	} `json:"result"`
}
