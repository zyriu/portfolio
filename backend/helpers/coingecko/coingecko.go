package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/misc"
)

const apiBaseUrl = "https://api.coingecko.com/api"

func FetchSimplePrices(ctx context.Context, gristRecords grist.Records) (map[string]float64, error) {
	prices := make(map[string]float64)

	coingeckoIDs := grist.ExtractColumnDataFromRecords[string](gristRecords, "Coingecko_ID")
	coingeckoIDs = slices.DeleteFunc(coingeckoIDs, func(v string) bool { return v == "" })

	const chunkSize = 100
	for i := 0; i < len(coingeckoIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(coingeckoIDs) {
			end = len(coingeckoIDs)
		}

		chunk := coingeckoIDs[i:end]
		q := url.Values{}
		q.Set("ids", strings.Join(chunk, ","))
		q.Set("vs_currencies", "usd")
		u := apiBaseUrl + "/v3/simple/price?" + q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return prices, err
		}

		resp := misc.DoWithRetry(ctx, req)
		if resp.Err != nil {
			return nil, resp.Err
		}

		respMap := make(map[string]map[string]float64)
		if err := json.Unmarshal(resp.Body, &respMap); err != nil {
			return nil, fmt.Errorf("decode response: %w", err)
		}

		for coin, m := range respMap {
			if usd, ok := m["usd"]; ok {
				prices[coin] = usd
			}
		}
	}

	return prices, nil
}
