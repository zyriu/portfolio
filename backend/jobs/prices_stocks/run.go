package prices_stocks

import (
	"context"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/twelveData"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Initializing TwelveData client...")
	f, err := twelveData.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	query := fmt.Sprintf("fields=%s", "Ticker")

	updateStatus("Fetching stock tickers from Grist Positions_TradFi_ table...")
	records, err := g.GetRecords(ctx, "Positions_TradFi_", query)
	if err != nil {
		return err
	}

	tickersList := grist.ExtractColumnDataFromRecords[string](records, "Ticker")
	updateStatus(fmt.Sprintf("Found %d stock ticker(s): %v", len(tickersList), tickersList))

	updateStatus("Fetching current prices from TwelveData API...")
	prices, err := f.GetBatchPrices(ctx, tickersList)
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("✓ Received %d price(s) from TwelveData", len(prices)))
	updateStatus("Processing stock price updates...")
	var upserts []grist.Upsert
	for _, price := range prices {
		updateStatus(fmt.Sprintf("  %s: $%s", price.Ticker, price.Price))
		upserts = append(upserts, grist.Upsert{
			Require: map[string]any{"Ticker": price.Ticker},
			Fields:  map[string]any{"Price": price.Price},
		})
	}

	if len(prices) < len(tickersList) {
		updateStatus(fmt.Sprintf("⚠️  Warning: Only received %d/%d prices", len(prices), len(tickersList)))
	}

	if len(upserts) > 0 {
		updateStatus(fmt.Sprintf("Upserting %d stock prices to Grist...", len(upserts)))
		if err := g.UpsertRecords(ctx, "Positions_TradFi_", upserts, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("✓ Successfully updated %d stock prices", len(upserts)))
	} else {
		updateStatus("⚠️  No stock prices to update")
	}

	return nil
}
