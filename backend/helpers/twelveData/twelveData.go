package twelveData

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

const apiBaseUrl = "https://api.twelvedata.com"

type TwelveData struct {
	apiKey string
}

type Price struct {
	Ticker string `json:"ticker,omitempty"`
	Price  string `json:"price"`
}

func InitiateClient() (TwelveData, error) {
	var twelveData TwelveData

	// Load settings
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return twelveData, fmt.Errorf("failed to load settings: %v", err)
	}

	twelveData.apiKey = settingsData.Settings.Stocks.TwelveDataAPIKey
	return twelveData, nil
}

func (f *TwelveData) GetBatchPrices(ctx context.Context, tickers []string) ([]Price, error) {
	endpoint := apiBaseUrl + "/price"
	q := url.Values{}
	q.Set("symbol", strings.Join(tickers, ","))
	q.Set("apikey", f.apiKey)
	u := endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	resp := misc.DoWithRetry(ctx, req)
	if resp.Err != nil {
		return nil, resp.Err
	}

	// Handle single ticker vs multiple tickers response format
	// Single ticker: {"price":"456.56000"}
	if len(tickers) == 1 {
		var price Price
		if err := json.Unmarshal(resp.Body, &price); err != nil {
			return nil, fmt.Errorf("failed to unmarshal single ticker price: %v", err)
		}

		price.Ticker = tickers[0]
		return []Price{price}, nil
	}

	// Multiple tickers: {"TSLA":{"price":"456.56000"},"AAPL":{"price":"270.37000"}}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resp.Body, &raw); err != nil {
		return nil, err
	}

	prices := make([]Price, 0, len(tickers))
	for _, ticker := range tickers {
		v, ok := raw[ticker]
		if !ok || len(v) == 0 {
			continue
		}

		var price Price
		if err := json.Unmarshal(v, &price); err != nil {
			return nil, err
		}

		price.Ticker = ticker
		prices = append(prices, price)
	}

	return prices, nil
}
