package pendle_markets

import (
	"context"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/pendle"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Initializing Pendle client...")
	p, err := pendle.InitiateClient()
	if err != nil {
		return err
	}

	var (
		assets  pendle.Assets
		markets pendle.Markets
		prices  grist.Prices
		tokens  grist.Tokens
	)

	updateStatus("Fetching data in parallel (assets, markets, prices, tokens)...")
	errGroup, c := errgroup.WithContext(ctx)
	misc.Go(errGroup, c, p.FetchAssets, &assets)
	misc.Go(errGroup, c, p.FetchAllMarkets, &markets)
	misc.Go(errGroup, c, g.FetchPrices, &prices)
	misc.Go(errGroup, c, g.FetchTokens, &tokens)

	if err := errGroup.Wait(); err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("✓ Fetched assets: %d, markets: %d, prices: %d, tokens: %d",
		len(assets.Assets), len(markets.Markets), len(prices), len(tokens)))

	updateStatus(fmt.Sprintf("Processing %d Pendle markets...", len(markets.Markets)))
	records := generatePendleRecords(markets)

	if len(records) == 0 {
		updateStatus("No Pendle market records to sync")
		return nil
	}

	updateStatus(fmt.Sprintf("Upserting %d market records to Grist...", len(records)))
	if err := g.UpsertRecords(ctx, "Pendle", records, grist.UpsertOpts{}); err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("✓ Successfully synced %d Pendle markets", len(records)))
	return nil
}
