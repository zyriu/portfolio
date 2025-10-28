package update_kraken

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/kraken"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func updateBalances(ctx context.Context, k kraken.Kraken, g grist.Grist) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Fetching balances from Kraken API...")
	balances, err := k.GetBalances(context.Background())
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("Found %d raw balance entries", len(balances)))
	var upserts []grist.Upsert

	totals := make(map[string]float64)

	updateStatus("Normalizing and aggregating balances...")
	for rawTicker, amountStr := range balances {
		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			return err
		}

		ticker := k.NormalizeTicker(rawTicker)
		totals[ticker] += amount
		updateStatus(fmt.Sprintf("  %s (%s): %.8f", ticker, rawTicker, amount))
	}

	updateStatus(fmt.Sprintf("Creating %d position records...", len(totals)))
	for ticker, total := range totals {
		upserts = append(upserts, grist.Upsert{
			Require: map[string]any{
				"Ticker": ticker,
				"Wallet": "Kraken",
			},
			Fields: map[string]any{
				"Amount":     total,
				"Asset_Type": token.IsStableOrVolatile(ticker),
			},
		})
	}

	if len(upserts) > 0 {
		updateStatus(fmt.Sprintf("Upserting %d positions to Grist...", len(upserts)))
		if err := g.UpsertRecords(ctx, "Positions_Crypto_", upserts, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("✓ Successfully synced %d positions", len(upserts)))
	} else {
		updateStatus("No positions to sync")
	}

	return nil
}
