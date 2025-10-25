package update_pendle_positions

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zyriu/portfolio/backend/helpers/chain"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/pendle"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func appendOpenedPositions(wallet string, yieldPositions []grist.Upsert, missingTickers []grist.Upsert,
	positions []pendle.OpenOrClosedPosition, markets pendle.MarketsMap, assets pendle.AssetsMap, prices grist.Prices,
	tokens grist.Tokens) ([]grist.Upsert, []grist.Upsert) {

	// Track seen missing tickers to avoid duplicates
	seenMissingTickers := make(map[string]bool)

	createRecord := func(c string, name string, expiry string, underlying string, exposure string, IY float64,
		poolType string, position pendle.TokenPosition) grist.Upsert {

		claimable := ""
		for _, claim := range position.ClaimTokenAmounts {
			if t, ok := tokens[claim.Token]; ok && claim.Amount != "0" {
				amount, _ := decimal.NewFromString(claim.Amount)
				decimalFactor := decimal.New(1, t.Decimal)
				normalizedAmount := amount.Div(decimalFactor)
				priceDec := decimal.NewFromFloat(prices[t.Ticker])

				toAdd, _ := decimal.NewFromString(claimable)
				claimable = normalizedAmount.Mul(priceDec).Add(toAdd).String()
			} else {
				split := strings.SplitN(claim.Token, "-", 2)
				chainID, _ := strconv.Atoi(split[0])
				chain := chain.IDToName(chainID)
				address := split[1]
				key := chain + ":" + address

				// Only add if we haven't seen this token before
				if !seenMissingTickers[key] {
					seenMissingTickers[key] = true
					missingTickers = append(missingTickers, grist.Upsert{
						Require: map[string]any{
							"Chain":   chain,
							"Address": address,
						},
						Fields: map[string]any{},
					})
				}
			}
		}

		return grist.Upsert{
			Require: map[string]any{
				"Wallet":   wallet,
				"Chain":    c,
				"Protocol": "Pendle",
				"Name":     name,
				"Expiry":   expiry,
			},
			Fields: map[string]any{
				"Underlying":     underlying,
				"Asset_Type":     token.IsStableOrVolatile(underlying),
				"Exposure":       exposure,
				"Pool_Type":      poolType,
				"Current_Value":  position.Valuation,
				"IY":             IY,
				"Claimable_USD_": claimable,
			},
		}
	}

	for _, position := range positions {
		// Check if market data exists (may be missing for matured/expired positions)
		market, marketExists := markets[position.MarketID]
		if !marketExists {
			// Skip matured positions that are no longer in the markets API
			continue
		}

		// Check if PT and SY assets exist
		ptAsset, ptExists := assets[market.PT]
		syAsset, syExists := assets[market.SY]
		if !ptExists || !syExists {
			// Skip positions with missing asset data
			continue
		}

		c := chain.IDToName(market.ChainID)
		IY := market.Details.ImpliedAPY

		re := regexp.MustCompile(`\s*\(([^)]+)\)$`)
		s := ptAsset.Name
		name := re.ReplaceAllString(s, "")

		// Extract asset from parentheses, or use the full name as fallback
		asset := s
		if matches := re.FindStringSubmatch(s); len(matches) > 1 {
			asset = matches[1]
		}

		exposure := syAsset.Name

		t, _ := time.Parse("2Jan2006", regexp.MustCompile(`(\d{1,2}[A-Z]{3}\d{4})`).FindString(ptAsset.Symbol))
		expiry := t.Format("02 Jan 2006")

		if position.LP.Valuation > 0 {
			name = strings.Replace(name, "PT", "LP", 1)
			yieldPositions = append(yieldPositions, createRecord(c, name, expiry, asset, exposure, IY, "LP", position.LP))
		}
		if position.PT.Valuation > 0 {
			yieldPositions = append(yieldPositions, createRecord(c, name, expiry, asset, exposure, IY, "PT", position.PT))
		}
		if position.YT.Valuation > 0 {
			name = strings.Replace(name, "PT", "YT", 1)
			yieldPositions = append(yieldPositions, createRecord(c, name, expiry, asset, exposure, IY, "YT", position.YT))
		}
	}

	return yieldPositions, missingTickers
}
