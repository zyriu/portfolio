package update_pendle_positions

import (
	"context"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/pendle"
	"github.com/zyriu/portfolio/backend/helpers/settings"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Loading settings...")
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return err
	}

	wallets := settings.GetPendleWallets(settingsData)

	// If no wallets configured, return gracefully
	if len(wallets) == 0 {
		updateStatus("No Pendle wallets configured")
		return nil
	}

	updateStatus(fmt.Sprintf("Found %d wallet(s) to check for Pendle positions", len(wallets)))

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

	updateStatus("Building assets and markets lookup maps...")
	assetsMap := pendle.GenerateAssetsMap(assets)
	marketsMap := pendle.GenerateMarketsMap(markets)
	updateStatus(fmt.Sprintf("✓ Built maps: %d assets, %d markets", len(assetsMap), len(marketsMap)))

	var yieldPositions []grist.Upsert
	var missingTickers []grist.Upsert

	for i, wallet := range wallets {
		updateStatus(fmt.Sprintf("Checking wallet %d/%d: %s (%s)", i+1, len(wallets), wallet.Label, wallet.Address))
		userPositions, err := p.GetUserPositions(ctx, wallet.Address)
		if err != nil {
			return err
		}

		updateStatus(fmt.Sprintf("[%s] Found %d position groups", wallet.Label, len(userPositions.Positions)))
		totalOpenPositions := 0.0
		for chainIdx, positions := range userPositions.Positions {
			if positions.TotalOpen == 0 {
				updateStatus(fmt.Sprintf("[%s]   Chain group %d: no open positions", wallet.Label, chainIdx+1))
				continue
			}

			updateStatus(fmt.Sprintf("[%s]   Chain group %d: %.0f open position(s)", wallet.Label, chainIdx+1, positions.TotalOpen))
			totalOpenPositions += positions.TotalOpen
			beforeCount := len(yieldPositions)
			yieldPositions, missingTickers = appendOpenedPositions(wallet.Label, yieldPositions, missingTickers,
				positions.OpenPositions, marketsMap, assetsMap, prices, tokens)
			addedCount := len(yieldPositions) - beforeCount
			updateStatus(fmt.Sprintf("[%s]   Processed %d position(s)", wallet.Label, addedCount))
		}
		updateStatus(fmt.Sprintf("[%s] ✓ Total open positions: %.0f", wallet.Label, totalOpenPositions))
	}

	if len(yieldPositions) > 0 {
		updateStatus(fmt.Sprintf("Upserting %d yield positions to Grist...", len(yieldPositions)))
		if err := g.UpsertRecords(ctx, "Yield", yieldPositions, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("✓ Successfully synced %d yield positions", len(yieldPositions)))
	} else {
		updateStatus("No yield positions found")
	}

	if len(missingTickers) > 0 {
		updateStatus(fmt.Sprintf("Adding %d missing tickers to Tokens table...", len(missingTickers)))
		if err := g.UpsertRecords(ctx, "Tokens", missingTickers, grist.UpsertOpts{}); err != nil {
			return err
		}
	}

	return nil
}
