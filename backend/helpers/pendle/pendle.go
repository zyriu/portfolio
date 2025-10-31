package pendle

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/zyriu/portfolio/backend/helpers/misc"
)

const apiBaseUrl = "https://api-v2.pendle.finance/core"

func InitiateClient() (Pendle, error) {
	var pendle Pendle

	return pendle, nil
}

func (p *Pendle) FetchAllMarkets(ctx context.Context) (Markets, error) {
	var markets Markets

	endpoint := "/v1/markets/all?isActive=true"
	u := apiBaseUrl + endpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	endpoint := "/v1/assets/all"
	u := apiBaseUrl + endpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	u := fmt.Sprintf("%s/v2/%d/markets/%s/data", apiBaseUrl, chainID, address)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

	endpoint := apiBaseUrl + "/v1/dashboard/positions/database"
	u := fmt.Sprintf("%s/%s", endpoint, user)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
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

func (p *Pendle) GetUserTransactions(ctx context.Context, user string, from time.Time) (UserTransactions, error) {
	var userTransactions UserTransactions

	endpoint := apiBaseUrl + "/v1/pnl/transactions"
	q := url.Values{}
	q.Set("user", user)
	q.Set("fromTimestamp", from.Format(time.RFC3339))
	u := endpoint + "?" + q.Encode()

	log.Println("u", u)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return userTransactions, err
	}

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return userTransactions, fmt.Errorf("pendle.GetUserTransactions: request failed: %w", resp.Err)
	}

	if resp.Status < 200 || resp.Status >= 300 {
		return userTransactions, fmt.Errorf("pendle.GetUserTransactions: unexpected status %d: %s", resp.Status, string(resp.Body))
	}

	if err := json.Unmarshal(resp.Body, &userTransactions); err != nil {
		return userTransactions, fmt.Errorf("pendle.GetUserTransactions: decode response: %w", err)
	}

	return userTransactions, nil
}
