package coingecko

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Allowlist for common blue-chips (optional but handy).
var bluechip = map[string]string{
	"btc":  "bitcoin",
	"eth":  "ethereum",
	"sol":  "solana",
	"usdc": "usd-coin",
	"usdt": "tether",
	"bnb":  "binancecoin",
	"xrp":  "ripple",
	"ada":  "cardano",
	"doge": "dogecoin",
}

// Fetch /coins/list?include_platform=true once.
func FetchCoinsListWithPlatforms(ctx context.Context) ([]CGListCoin, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET",
		"https://api.coingecko.com/api/v3/coins/list?include_platform=true", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var list []CGListCoin
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	return list, nil
}

func BuildIndex(coins []CGListCoin) *Index {
	idx := &Index{
		BySymbol: make(map[string][]CGListCoin),
		ByName:   make(map[string][]CGListCoin),
	}
	for _, c := range coins {
		sym := strings.ToLower(c.Symbol)
		name := strings.ToLower(c.Name)
		idx.BySymbol[sym] = append(idx.BySymbol[sym], c)
		idx.ByName[name] = append(idx.ByName[name], c)
	}
	return idx
}

// When collisions remain, pick best by market cap.
// We batch up to 250 ids per call to /coins/markets.
func PickByMarketCap(ctx context.Context, ids []string) (string, error) {
	if len(ids) == 0 {
		return "", nil
	}
	// Build the query with comma-separated ids (<=250 per request).
	q := url.Values{}
	q.Set("vs_currency", "usd")
	q.Set("ids", strings.Join(ids, ","))
	q.Set("order", "market_cap_desc")
	q.Set("per_page", strconv.Itoa(len(ids)))
	q.Set("page", "1")

	endpoint := "https://api.coingecko.com/api/v3/coins/markets?" + q.Encode()
	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var mkt []CGMarketCoin
	if err := json.NewDecoder(resp.Body).Decode(&mkt); err != nil {
		return "", err
	}
	if len(mkt) == 0 {
		return "", nil
	}

	// Choose the one with the best (lowest) MarketCapRank; if nil, use MarketCap fallback.
	bestID := mkt[0].ID
	bestRank := 1 << 30
	bestCap := -1.0

	for _, c := range mkt {
		rank := 1 << 29
		if c.MarketCapRank != nil {
			rank = *c.MarketCapRank
		}
		cap := -1.0
		if c.MarketCap != nil {
			cap = *c.MarketCap
		}
		// Prefer valid rank; if both have rank, pick the smaller.
		// If ranks missing or equal, prefer larger market cap.
		if rank < bestRank || (rank == bestRank && cap > bestCap) {
			bestRank = rank
			bestCap = cap
			bestID = c.ID
		}
	}
	return bestID, nil
}

// Resolve a single record to a CoinGecko ID using heuristics.
func ResolveID(ctx context.Context, idx *Index, r Record) (string, error) {
	ticker := strings.ToLower(strings.TrimSpace(r.Ticker))
	if ticker == "" {
		return "", nil
	}

	// 0) Blue-chip fast path.
	if id, ok := bluechip[ticker]; ok {
		return id, nil
	}

	candidates := idx.BySymbol[ticker]
	if len(candidates) == 0 {
		// As a last resort, try an exact name match if available.
		if r.Name != "" {
			name := strings.ToLower(strings.TrimSpace(r.Name))
			if byName, ok := idx.ByName[name]; ok && len(byName) == 1 {
				return byName[0].ID, nil
			}
		}
		return "", nil
	}

	if len(candidates) == 1 {
		return candidates[0].ID, nil
	}

	// 1) If we have exact name, try to filter candidates by name.
	if r.Name != "" {
		name := strings.ToLower(strings.TrimSpace(r.Name))
		filtered := make([]CGListCoin, 0, len(candidates))
		for _, c := range candidates {
			if strings.ToLower(c.Name) == name {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) == 1 {
			return filtered[0].ID, nil
		}
		if len(filtered) > 1 {
			candidates = filtered
		}
	}

	// 2) If we have chain+contract, match by platforms precisely.
	if r.Chain != "" && r.Contract != "" {
		chain := strings.ToLower(strings.TrimSpace(r.Chain))
		contract := strings.ToLower(strings.TrimSpace(r.Contract))
		for _, c := range candidates {
			for ch, addr := range c.Platforms {
				if strings.ToLower(ch) == chain && strings.EqualFold(strings.ToLower(addr), contract) {
					return c.ID, nil
				}
			}
		}
	}

	// 3) Still ambiguous → pick the largest by market cap (best rank).
	ids := make([]string, 0, len(candidates))
	for _, c := range candidates {
		ids = append(ids, c.ID)
	}
	return PickByMarketCap(ctx, ids)
}
