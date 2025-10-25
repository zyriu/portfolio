package update_lighter

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/lighter"
	"github.com/zyriu/portfolio/backend/helpers/settings"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func Run(ctx context.Context, args ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Loading settings...")
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return err
	}

	wallets := settings.GetLighterWallets(settingsData)

	// If no wallets configured, return gracefully
	if len(wallets) == 0 {
		updateStatus("No wallets configured")
		return nil
	}

	updateStatus(fmt.Sprintf("Found %d wallet(s) to sync", len(wallets)))

	updateStatus("Initializing Lighter client...")
	l, err := lighter.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	var upserts []grist.Upsert

	for i, wallet := range wallets {
		updateStatus(fmt.Sprintf("Processing wallet %d/%d: %s", i+1, len(wallets), wallet.Label))
		updateStatus(fmt.Sprintf("[%s] Fetching accounts from Lighter API for address %s...", wallet.Label, wallet.Address))
		accounts, err := l.GetAccounts(ctx, wallet.Address)
		if err != nil {
			return err
		}

		updateStatus(fmt.Sprintf("[%s] Found %d account(s)", wallet.Label, len(accounts)))
		totals := make(map[string]float64)
		totalPositions := 0

		for accountIdx, account := range accounts {
			updateStatus(fmt.Sprintf("[%s] Processing account %d/%d with %d position(s)...", wallet.Label, accountIdx+1, len(accounts), len(account.Positions)))
			for _, position := range account.Positions {
				amount, err := strconv.ParseFloat(position.PositionValue, 64)
				if err != nil {
					return err
				}

				totals[position.Symbol] += amount
				totalPositions++
				updateStatus(fmt.Sprintf("[%s]   %s: %.4f", wallet.Label, position.Symbol, amount))
			}
		}

		updateStatus(fmt.Sprintf("[%s] Aggregated %d positions into %d unique tickers", wallet.Label, totalPositions, len(totals)))
		for ticker, total := range totals {
			updateStatus(fmt.Sprintf("[%s] Adding to upsert payload: %s %f", wallet.Label, ticker, total))
			upserts = append(upserts, grist.Upsert{
				Require: map[string]any{
					"Ticker": ticker,
					"Wallet": wallet.Label,
					"Chain":  "Lighter",
				},
				Fields: map[string]any{
					"Amount":     total,
					"Asset_Type": token.IsStableOrVolatile(ticker),
				},
			})
		}
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
