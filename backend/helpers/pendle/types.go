package pendle

import "time"

type Pendle struct{}

type Asset struct {
	Name     string   `json:"name"`
	Decimals int      `json:"decimals"`
	Address  string   `json:"address"`
	Symbol   string   `json:"symbol"`
	Tags     []string `json:"tags"`
	Expiry   string   `json:"expiry"`
	ProIcon  string   `json:"proIcon"`
	ChainID  int      `json:"chainId"`
}

type Assets struct {
	Assets []Asset `json:"assets"`
}

type AssetsMap map[string]Asset

type claimTokenAmount struct {
	Token  string `json:"token"`
	Amount string `json:"amount"`
}

type Market struct {
	Name            string `json:"name"`
	Address         string `json:"address"`
	Expiry          string `json:"expiry"`
	PT              string `json:"pt"`
	YT              string `json:"yt"`
	SY              string `json:"sy"`
	UnderlyingAsset string `json:"underlyingAsset"`
	Details         struct {
		Liquidity     float64 `json:"liquidity"`
		PendleAPY     float64 `json:"pendleApy"`
		ImpliedAPY    float64 `json:"impliedApy"`
		FeeRate       float64 `json:"feeRate"`
		AggregatedAPY float64 `json:"aggregatedApy"`
		MaxBoostedAPY float64 `json:"maxBoostedApy"`
	} `json:"details"`
	IsNew       bool     `json:"isNew"`
	IsPrime     bool     `json:"isPrime"`
	Timestamp   string   `json:"timestamp"`
	CategoryIDs []string `json:"categoryIds"`
	ChainID     int      `json:"chainId"`
}

type Markets struct {
	Markets []Market `json:"markets"`
}

type MarketsMap map[string]Market

type MarketDetailedData struct {
	Timestamp time.Time `json:"timestamp"`
	Liquidity struct {
		Usd float64 `json:"usd"`
		Acc float64 `json:"acc"`
	} `json:"liquidity"`
	TradingVolume struct {
		Usd float64 `json:"usd"`
	} `json:"tradingVolume"`
	TotalTvl struct {
		Usd float64 `json:"usd"`
	} `json:"totalTvl"`
	UnderlyingInterestApy     float64 `json:"underlyingInterestApy"`
	UnderlyingRewardApy       float64 `json:"underlyingRewardApy"`
	UnderlyingApy             float64 `json:"underlyingApy"`
	ImpliedApy                float64 `json:"impliedApy"`
	YtFloatingApy             float64 `json:"ytFloatingApy"`
	SwapFeeApy                float64 `json:"swapFeeApy"`
	VoterApy                  float64 `json:"voterApy"`
	PtDiscount                float64 `json:"ptDiscount"`
	PendleApy                 float64 `json:"pendleApy"`
	ArbApy                    float64 `json:"arbApy"`
	LpRewardApy               float64 `json:"lpRewardApy"`
	AggregatedApy             float64 `json:"aggregatedApy"`
	MaxBoostedApy             float64 `json:"maxBoostedApy"`
	EstimatedDailyPoolRewards []struct {
		Asset struct {
			ID          string  `json:"id"`
			ChainID     int     `json:"chainId"`
			Address     string  `json:"address"`
			Symbol      string  `json:"symbol"`
			Decimals    int     `json:"decimals"`
			AccentColor *string `json:"accentColor"`
			Price       struct {
				Usd float64 `json:"usd"`
			} `json:"price"`
			PriceUpdatedAt time.Time `json:"priceUpdatedAt"`
			Name           string    `json:"name"`
		} `json:"asset"`
		Amount float64 `json:"amount"`
	} `json:"estimatedDailyPoolRewards"`
	TotalPt           float64 `json:"totalPt"`
	TotalSy           float64 `json:"totalSy"`
	TotalLp           float64 `json:"totalLp"`
	TotalActiveSupply float64 `json:"totalActiveSupply"`
	AssetPriceUsd     float64 `json:"assetPriceUsd"`
}

type TokenPosition struct {
	Balance           string             `json:"balance"`
	ActiveBalance     string             `json:"activeBalance"`
	Valuation         float64            `json:"valuation"`
	ClaimTokenAmounts []claimTokenAmount `json:"claimTokenAmounts"`
}

type OpenOrClosedPosition struct {
	MarketID string        `json:"marketId"`
	PT       TokenPosition `json:"pt"`
	YT       TokenPosition `json:"yt"`
	LP       TokenPosition `json:"lp"`
}

type SyPosition struct {
	SyID              string             `json:"syId"`
	Balance           string             `json:"balance"`
	ClaimTokenAmounts []claimTokenAmount `json:"claimTokenAmounts"`
}

type UserPositions struct {
	Positions []struct {
		ChainID         int                    `json:"chainId"`
		TotalOpen       float64                `json:"totalOpen"`
		TotalClosed     float64                `json:"totalClosed"`
		TotalSy         float64                `json:"totalSy"`
		OpenPositions   []OpenOrClosedPosition `json:"openPositions"`
		ClosedPositions []OpenOrClosedPosition `json:"closedPositions"`
		SyPositions     []SyPosition           `json:"syPositions"`
		UpdatedAt       string                 `json:"updatedAt"`
		ErrorMessage    string                 `json:"errorMessage"`
	} `json:"positions"`
}
