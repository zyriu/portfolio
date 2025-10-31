package pendle_user_positions

import (
	"context"
	"fmt"
	"log"
	"time"

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

	// Update missing tickers to Tokens table to be able to compute claimable amounts
	var missingTickers map[string][]string

	for i, wallet := range wallets {
		updateStatus(fmt.Sprintf("Checking wallet %d/%d: %s (%s)", i+1, len(wallets), wallet.Label, wallet.Address))
		userPositions, err := p.GetUserPositions(ctx, wallet.Address)
		if err != nil {
			return err
		}

		updateStatus(fmt.Sprintf("[%s] Found %d position groups", wallet.Label, len(userPositions.Positions)))

		var earliestTimestamp time.Time
		for chainIdx, positions := range userPositions.Positions {
			if positions.TotalOpen == 0 {
				updateStatus(fmt.Sprintf("[%s]   Chain group %d: no open positions", wallet.Label, chainIdx+1))
				continue
			}

			var processedPositions []grist.Upsert
			processedPositions, earliestTimestamp, missingTickers = processPositions(wallet.Label, positions.OpenPositions, earliestTimestamp, missingTickers, assetsMap, marketsMap, prices, tokens)
			yieldPositions = append(yieldPositions, processedPositions...)
			updateStatus(fmt.Sprintf("[%s]   Processed %d position(s)", wallet.Label, len(processedPositions)))
		}

		updateStatus(fmt.Sprintf("[%s] Fetching transactions since %s", wallet.Label, earliestTimestamp.UTC().Format(time.RFC3339)))
		userTransactions, err := p.GetUserTransactions(ctx, wallet.Address, earliestTimestamp)
		if err != nil {
			return err
		}

		updateStatus(fmt.Sprintf("[%s] Found %d transaction(s)", wallet.Label, userTransactions.Total))
		for _, transaction := range userTransactions.Results {
			log.Println("transaction", transaction)
		}
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
		missingTickersRecords := make([]grist.Upsert, 0, len(missingTickers))
		for chain, addresses := range missingTickers {
			for _, address := range addresses {
				missingTickersRecords = append(missingTickersRecords, grist.Upsert{
					Require: map[string]any{
						"Chain":   chain,
						"Address": address,
					},
					Fields: map[string]any{},
				})
			}
		}

		updateStatus(fmt.Sprintf("Adding %d missing tickers to Tokens table...", len(missingTickersRecords)))
		if err := g.UpsertRecords(ctx, "Tokens", missingTickersRecords, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("✓ Successfully synced %d missing tickers", len(missingTickersRecords)))
	} else {
		updateStatus("No missing tickers found")
	}

	return nil
}
