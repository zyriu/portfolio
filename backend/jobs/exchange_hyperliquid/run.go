package exchange_hyperliquid

import (
	"context"
	"fmt"
	"sync"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/hyperliquid"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
)

func Run(ctx context.Context, args ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Loading settings...")
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return err
	}

	wallets := settings.GetHyperliquidWallets(settingsData)

	// If no wallets configured, return gracefully
	if len(wallets) == 0 {
		updateStatus("No wallets configured")
		return nil
	}

	updateStatus(fmt.Sprintf("Found %d wallet(s) to sync", len(wallets)))

	updateStatus("Initializing Hyperliquid client...")
	h, err := hyperliquid.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	// Process each wallet
	for i, wallet := range wallets {
		updateStatus(fmt.Sprintf("Processing wallet %d/%d: %s", i+1, len(wallets), wallet.Label))
		// Execute updates concurrently for this wallet
		var wg sync.WaitGroup
		errChan := make(chan error, 2)
		statusChan := make(chan string, 10)

		// Status updater goroutine
		go func() {
			for status := range statusChan {
				updateStatus(status)
			}
		}()

		wg.Add(2)

		// Convert settings.UnifiedWallet to misc.Wallet
		miscWallet := misc.Wallet{
			Label:   wallet.Label,
			Address: wallet.Address,
			Type:    wallet.Type,
		}

		// Update balances concurrently
		go func() {
			defer wg.Done()
			statusChan <- fmt.Sprintf("[%s] Fetching balances...", wallet.Label)
			if err := updateBalances(ctx, h, g, miscWallet); err != nil {
				errChan <- fmt.Errorf("update balances for %s: %w", wallet.Label, err)
				return
			}
			statusChan <- fmt.Sprintf("[%s] ✓ Balances synced", wallet.Label)
		}()

		// Update trades concurrently
		go func() {
			defer wg.Done()
			statusChan <- fmt.Sprintf("[%s] Fetching trades...", wallet.Label)
			if err := updateTrades(ctx, h, g, miscWallet); err != nil {
				errChan <- fmt.Errorf("update trades for %s: %w", wallet.Label, err)
				return
			}
			statusChan <- fmt.Sprintf("[%s] ✓ Trades synced", wallet.Label)
		}()

		wg.Wait()
		close(errChan)
		close(statusChan)

		// Check for errors
		for err := range errChan {
			if err != nil {
				return err
			}
		}
	}

	updateStatus(fmt.Sprintf("✓ All %d wallets synced successfully", len(wallets)))
	return nil
}
