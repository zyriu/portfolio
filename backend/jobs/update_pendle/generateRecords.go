package update_pendle

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zyriu/portfolio/backend/helpers/chain"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/pendle"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func appendOpenedPositions(wallet string, yieldPositions []grist.Upsert, missingTickers []grist.Upsert,
	positions []pendle.OpenOrClosedPosition, markets pendle.MarketsMap, assets pendle.AssetsMap, prices grist.Prices,
	tokens grist.Tokens) ([]grist.Upsert, []grist.Upsert) {

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
				missingTickers = append(missingTickers, grist.Upsert{
					Require: map[string]any{
						"Chain":   chain.IDToName(chainID),
						"Address": split[1],
					},
					Fields: map[string]any{},
				})
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
				"Asset_Type":     token.GetAssetType(underlying),
				"Exposure":       exposure,
				"Pool_Type":      poolType,
				"Current_Value":  position.Valuation,
				"IY":             IY,
				"Claimable_USD_": claimable,
			},
		}
	}

	for _, position := range positions {
		market := markets[position.MarketID]
		c := chain.IDToName(market.ChainID)
		IY := markets[position.MarketID].Details.ImpliedAPY

		re := regexp.MustCompile(`\s*\(([^)]+)\)$`)
		s := assets[market.PT].Name
		name := re.ReplaceAllString(s, "")
		asset := re.FindStringSubmatch(s)[1]
		exposure := assets[market.SY].Name

		t, _ := time.Parse("2Jan2006", regexp.MustCompile(`(\d{1,2}[A-Z]{3}\d{4})`).FindString(assets[market.PT].Symbol))
		expiry := t.Format("02 Jan 2006")

		if position.LP.Valuation > 0 {
			yieldPositions = append(yieldPositions, createRecord(c, name, expiry, asset, exposure, IY, "LP", position.LP))
		}
		if position.PT.Valuation > 0 {
			yieldPositions = append(yieldPositions, createRecord(c, name, expiry, asset, exposure, IY, "PT", position.PT))
		}
		if position.YT.Valuation > 0 {
			yieldPositions = append(yieldPositions, createRecord(c, name, expiry, asset, exposure, IY, "YT", position.YT))
		}
	}

	return yieldPositions, missingTickers
}

func generatePendleRecords(markets pendle.Markets) []grist.Upsert {
	records := make([]grist.Upsert, 0, len(markets.Markets))

	for _, market := range markets.Markets {
		var cats []string
		for _, id := range market.CategoryIDs {
			cats = append(cats, misc.Capitalize(id))
		}

		records = append(records, grist.Upsert{
			Require: map[string]any{
				"Chain":  chain.IDToName(market.ChainID),
				"Name":   market.Name,
				"Expiry": market.Expiry,
			},
			Fields: map[string]any{
				"Categories":  strings.Join(cats, ", "),
				"Is_New":      market.IsNew,
				"Is_Prime":    market.IsPrime,
				"Implied_APY": market.Details.ImpliedAPY,
				"Pendle_APY":  market.Details.PendleAPY,
				"Liquidity":   market.Details.Liquidity,
			},
		})
	}

	return records
}
