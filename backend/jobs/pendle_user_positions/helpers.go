package pendle_user_positions

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zyriu/portfolio/backend/helpers/chain"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/pendle"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func createRecord(
	wallet string,
	market pendle.Market,
	openPosition pendle.OpenOrClosedPosition,
	rawName string,
	exposure string,
	expiry string,
	prices grist.Prices,
	tokens grist.Tokens,
	missingTickers map[string][]string) (grist.Upsert, map[string][]string) {

	re := regexp.MustCompile(`\s*\(([^)]+)\)$`)
	name := re.ReplaceAllString(rawName, "")
	underlying := rawName
	if matches := re.FindStringSubmatch(rawName); len(matches) > 1 {
		underlying = matches[1]
	}

	poolType := "PT"
	position := openPosition.PT
	if openPosition.LP.Valuation > 0 {
		poolType = "LP"
		position = openPosition.LP
		name = strings.Replace(name, "PT", "LP", 1)
	} else if openPosition.YT.Valuation > 0 {
		poolType = "YT"
		position = openPosition.YT
		name = strings.Replace(name, "PT", "YT", 1)
	}

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

			if !slices.Contains(missingTickers[chain], address) {
				missingTickers[chain] = append(missingTickers[chain], address)
			}
		}
	}

	return grist.Upsert{
		Require: map[string]any{
			"Wallet":   wallet,
			"Chain":    chain.IDToName(market.ChainID),
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
			"IY":             market.Details.ImpliedAPY,
			"Claimable_USD_": claimable,
		},
	}, missingTickers
}

func processPositions(
	wallet string,
	positions []pendle.OpenOrClosedPosition,
	earliestTimestamp time.Time,
	missingTickers map[string][]string,
	assets pendle.AssetsMap,
	markets pendle.MarketsMap,
	prices grist.Prices,
	tokens grist.Tokens) ([]grist.Upsert, time.Time, map[string][]string) {

	processedPositions := make([]grist.Upsert, 0, len(positions))

	dateRe := regexp.MustCompile(`(\d{1,2}[A-Z]{3}\d{4})`)

	for _, position := range positions {
		// Skip matured positions that are no longer in the markets API
		market, marketExists := markets[position.MarketID]
		if !marketExists {
			continue
		}

		ptAsset, ptExists := assets[market.PT]
		syAsset, syExists := assets[market.SY]
		if !ptExists || !syExists {
			// Skip positions with missing asset data
			continue
		}

		if market.Timestamp != "" {
			if mt, perr := time.Parse(time.RFC3339, market.Timestamp); perr == nil && (earliestTimestamp.IsZero() || mt.Before(earliestTimestamp)) {
				earliestTimestamp = mt
			}
		}

		// Defensive: check FindString returns nonempty string; parse errors are ignored but ensure zero time isn't used for garbage
		dateStr := dateRe.FindString(ptAsset.Symbol)
		expiry := ""
		if dateStr != "" {
			if t, err := time.Parse("2Jan2006", dateStr); err == nil {
				expiry = t.Format("02 Jan 2006")
			}
		}

		var record grist.Upsert
		record, missingTickers = createRecord(wallet, market, position, ptAsset.Name, syAsset.Name, expiry, prices, tokens, missingTickers)
		processedPositions = append(processedPositions, record)
	}

	return processedPositions, earliestTimestamp, missingTickers
}
