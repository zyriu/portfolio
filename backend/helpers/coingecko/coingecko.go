package coingecko

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/misc"
)

const apiBaseUrl = "https://api.coingecko.com/api"

func FetchCoinMarkets(ctx context.Context, coingeckoIDs []string) ([]CoinMarket, error) {
	markets := make([]CoinMarket, 0)

	const chunkSize = 250
	for i := 0; i < len(coingeckoIDs); i += chunkSize {
		end := min(i+chunkSize, len(coingeckoIDs))

		chunk := coingeckoIDs[i:end]
		q := url.Values{}
		q.Set("ids", strings.Join(chunk, ","))
		q.Set("vs_currency", "usd")
		u := apiBaseUrl + "/v3/coins/markets?" + q.Encode()

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}

		resp := misc.DoWithRetry(ctx, req)
		if resp.Err != nil {
			return nil, fmt.Errorf("send request: %w", resp.Err)
		}

		var m []CoinMarket
		if err := json.Unmarshal(resp.Body, &m); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}

		markets = append(markets, m...)
	}

	return markets, nil
}
