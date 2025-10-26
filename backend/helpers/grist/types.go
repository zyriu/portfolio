package grist

type Grist struct {
	ApiKey string
	DocId  string
}

type Book map[string]BookEntry

type BookEntry struct {
	Exchange     string  `json:"exchange"`
	Ticker       string  `json:"Ticker"`
	Market       string  `json:"Market"`
	AssetType    string  `json:"Asset_Type"`
	AveragePrice float64 `json:"Average_Price"`
	PositionSize float64 `json:"Position_Size"`
	CostBasis    float64 `json:"Cost_Basis"`
}

type Price struct {
	Ticker      string  `json:"Ticker"`
	CoingeckoID string  `json:"Coingecko_ID"`
	Price       float64 `json:"Price"`
}

type Prices map[string]float64

type Record struct {
	RecordID int64          `json:"id"`
	Fields   map[string]any `json:"fields"`
}

type Records struct {
	Records []Record
}

type Token struct {
	Chain   string `json:"Chain"`
	Address string `json:"Address"`
	Ticker  string `json:"Ticker"`
	Decimal int32  `json:"Decimal"`
}

type Tokens map[string]Token

type Trade struct {
	Time             int64   `json:"Time"`
	OrderValue       float64 `json:"Order_Value"`
	Direction        string  `json:"Direction"`
	Exchange         string  `json:"Exchange"`
	Market           string  `json:"Market"`
	OrderType        string  `json:"Order_Type"`
	Price            float64 `json:"Price"`
	OrderSize        float64 `json:"Order_Size"`
	Ticker           string  `json:"Ticker"`
	Fee              float64 `json:"Fee"`
	FeeCurrency      string  `json:"Fee_Currency"`
	FeeUSD           float64 `json:"Fee_USD_"`
	PnL              float64 `json:"PnL"`
	TradeID          string  `json:"Trade_ID"`
	AggregatedTrades int     `json:"Aggregated_Trades"`
}

type UpsertOpts struct {
	AllowEmptyRequire bool
	OnMany            string
}

type Upsert struct {
	Require map[string]any `json:"require"`
	Fields  map[string]any `json:"fields"`
}

type upsertPayload struct {
	Records []Upsert `json:"records"`
}
