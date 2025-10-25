package coingecko

type CGListCoin struct {
	ID        string            `json:"id"`
	Symbol    string            `json:"symbol"`
	Name      string            `json:"name"`
	Platforms map[string]string `json:"platforms"` // chain -> contract address ("" for natives)
}

type CGMarketCoin struct {
	ID            string   `json:"id"`
	Symbol        string   `json:"symbol"`
	Name          string   `json:"name"`
	MarketCapRank *int     `json:"market_cap_rank"`
	MarketCap     *float64 `json:"market_cap"`
}

type Record struct {
	Ticker   string
	Name     string // optional if you have it
	Chain    string // e.g., "ethereum", "solana" (optional)
	Contract string // hex address for EVM, mint for Solana (optional)
}

type Index struct {
	BySymbol map[string][]CGListCoin
	ByName   map[string][]CGListCoin // keep slice in case of collisions on name too
}
