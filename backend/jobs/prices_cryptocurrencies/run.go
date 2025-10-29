package prices_cryptocurrencies

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/coingecko"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Fetching price records from Grist Prices table...")
	records, err := g.GetRecords(ctx, "Prices", "fields=Ticker,Coingecko_ID")
	if err != nil {
		return err
	}

	coingeckoIDs := grist.ExtractColumnDataFromRecords[string](records, "Coingecko_ID")
	coingeckoIDs = slices.DeleteFunc(coingeckoIDs, func(v string) bool { return v == "" })

	updateStatus("Fetching coin markets from CoinGecko API...")
	markets, err := coingecko.FetchCoinMarkets(ctx, coingeckoIDs)
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("✓ Received %d coin markets from CoinGecko", len(markets)))
	updateStatus("Processing and matching prices with tokens...")

	var upserts []grist.Upsert

	marketByID := make(map[string]coingecko.CoinMarket, len(markets))
	for _, m := range markets {
		marketByID[m.ID] = m
	}

	for _, r := range records.Records {
		raw, ok := r.Fields["Coingecko_ID"]
		if !ok || raw == nil {
			continue
		}

		coinID := strings.TrimSpace(fmt.Sprint(raw))
		if coinID == "" {
			continue
		}

		market, ok := marketByID[coinID]
		if !ok {
			continue
		}

		upserts = append(upserts, grist.Upsert{
			Require: map[string]any{
				"Coingecko_ID": coinID,
			},
			Fields: map[string]any{
				"Price":         market.CurrentPrice,
				"All_Time_High": market.ATH,
			},
		})
	}

	if len(upserts) > 0 {
		updateStatus(fmt.Sprintf("Upserting %d prices to Grist...", len(upserts)))
		if err := g.UpsertRecords(ctx, "Prices", upserts, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("✓ Successfully updated %d prices", len(upserts)))
	} else {
		updateStatus("⚠️  No prices to update")
	}

	return nil
}
