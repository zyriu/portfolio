package pendle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zyriu/portfolio/backend/helpers/misc"
)

const apiBaseUrl = "https://api-v2.pendle.finance/core"

func InitiateClient() (Pendle, error) {
	var pendle Pendle

	return pendle, nil
}

func (p *Pendle) FetchAllMarkets(ctx context.Context) (Markets, error) {
	var markets Markets

	path := "/v1/markets/all?isActive=true"
	endpoint := apiBaseUrl + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return markets, err
	}

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return markets, resp.Err
	}

	if err := json.Unmarshal(resp.Body, &markets); err != nil {
		return markets, fmt.Errorf("pendle.FetchAllMarkets: decode response: %w", err)
	}

	return markets, nil
}

func (p *Pendle) FetchAssets(ctx context.Context) (Assets, error) {
	var assets Assets

	path := "/v1/assets/all"
	endpoint := apiBaseUrl + path

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return assets, err
	}

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return assets, resp.Err
	}

	if err := json.Unmarshal(resp.Body, &assets); err != nil {
		return assets, fmt.Errorf("pendle.FetchAssets: decode response: %w", err)
	}

	return assets, nil
}

func (p *Pendle) GetDetailedMarketData(ctx context.Context, chainID int, address string) (MarketDetailedData, error) {
	var marketDetailedData MarketDetailedData

	endpoint := fmt.Sprintf("%s/v2/%d/markets/%s/data", apiBaseUrl, chainID, address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return marketDetailedData, err
	}

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return marketDetailedData, resp.Err
	}

	if err := json.Unmarshal(resp.Body, &marketDetailedData); err != nil {
		return marketDetailedData, fmt.Errorf("pendle.GetDetailedMarketData: decode response: %w", err)
	}

	return marketDetailedData, nil
}

func (p *Pendle) GetUserPositions(ctx context.Context, user string) (UserPositions, error) {
	var userPositions UserPositions

	path := "/v1/dashboard/positions/database"
	endpoint := fmt.Sprintf("%s%s/%s", apiBaseUrl, path, user)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return userPositions, err
	}

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return userPositions, resp.Err
	}

	if err := json.Unmarshal(resp.Body, &userPositions); err != nil {
		return userPositions, fmt.Errorf("pendle.GetUserPositions: decode response: %w", err)
	}

	return userPositions, nil
}
