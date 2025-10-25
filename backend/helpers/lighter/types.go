package lighter

type Lighter struct{}

type Position struct {
	MarketID               int    `json:"market_id"`
	Symbol                 string `json:"symbol"`
	InitialMarginFraction  string `json:"initial_margin_fraction"`
	OpenOrderCount         int    `json:"open_order_count"`
	PendingOrderCount      int    `json:"pending_order_count"`
	PositionTiedOrderCount int    `json:"position_tied_order_count"`
	Sign                   int    `json:"sign"`
	Position               string `json:"position"`
	AvgEntryPrice          string `json:"avg_entry_price"`
	PositionValue          string `json:"position_value"`
	UnrealizedPNL          string `json:"unrealized_pnl"`
	RealizedPNL            string `json:"realized_pnl"`
	LiquidationPrice       string `json:"liquidation_price"`
	MarginMode             int    `json:"margin_mode"`
	AllocatedMargin        string `json:"allocated_margin"`
}

type Account struct {
	Code                     int        `json:"code"`
	AccountType              int        `json:"account_type"`
	Index                    int        `json:"index"`
	L1Address                string     `json:"l1_address"`
	CancelAllTime            int        `json:"cancel_all_time"`
	TotalOrderCount          int        `json:"total_order_count"`
	TotalIsolatedOrderCount  int        `json:"total_isolated_order_count"`
	PendingOrderCount        int        `json:"pending_order_count"`
	AvailableBalance         string     `json:"available_balance"`
	Status                   int        `json:"status"`
	Collateral               string     `json:"collateral"`
	AccountIndex             int        `json:"account_index"`
	Name                     string     `json:"name"`
	Description              string     `json:"description"`
	CanInvite                bool       `json:"can_invite"`
	ReferralPointsPercentage string     `json:"referral_points_percentage"`
	Positions                []Position `json:"positions"`
	TotalAssetValue          string     `json:"total_asset_value"`
	CrossAssetValue          string     `json:"cross_asset_value"`
	Shares                   []any      `json:"shares"`
}

type AccountsResponse struct {
	Code     int       `json:"code"`
	Total    int       `json:"total"`
	Accounts []Account `json:"accounts"`
}

type SubAccount struct {
	Code                    int    `json:"code"`
	AccountType             int    `json:"account_type"`
	Index                   int    `json:"index"`
	L1Address               string `json:"l1_address"`
	CancelAllTime           int    `json:"cancel_all_time"`
	TotalOrderCount         int    `json:"total_order_count"`
	TotalIsolatedOrderCount int    `json:"total_isolated_order_count"`
	PendingOrderCount       int    `json:"pending_order_count"`
	AvailableBalance        string `json:"available_balance"`
	Status                  int    `json:"status"`
	Collateral              string `json:"collateral"`
}

type SubAccountsResponse struct {
	Code        int          `json:"code"`
	L1Address   string       `json:"l1_address"`
	SubAccounts []SubAccount `json:"sub_accounts"`
}
