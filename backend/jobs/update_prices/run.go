package update_prices

import (
	"context"
	"fmt"
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

	updateStatus("Fetching current prices from CoinGecko API...")
	prices, err := coingecko.FetchSimplePrices(ctx, records)
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("✓ Received %d prices from CoinGecko", len(prices)))
	updateStatus("Processing and matching prices with tokens...")
	var upserts []grist.Upsert
	skipped := 0
	seen := make(map[string]bool)
	for _, r := range records.Records {
		raw, ok := r.Fields["Coingecko_ID"]
		if !ok || raw == nil {
			skipped++
			continue
		}

		coinID := strings.TrimSpace(fmt.Sprint(raw))
		if coinID == "" {
			skipped++
			continue
		}

		price, ok := prices[coinID]
		if !ok {
			skipped++
			continue
		}

		// Check if we've already seen this CoinGecko_ID and upserted it
		if _, exists := seen[coinID]; exists {
			continue
		}
		seen[coinID] = true

		upserts = append(upserts, grist.Upsert{
			Require: map[string]any{
				"Coingecko_ID": coinID,
			},
			Fields: map[string]any{
				"Price": price,
			},
		})
	}

	if skipped > 0 {
		updateStatus(fmt.Sprintf("Skipped %d tokens (missing CoinGecko ID or price not available)", skipped))
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
