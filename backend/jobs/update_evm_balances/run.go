package update_evm_balances

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/coinstats"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

const chainsToUpsert = "ethereum, arbitrum-one, bsc, hyperevm, base"

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Loading settings...")
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %v", err)
	}

	wallets := settings.GetEVMWallets(settingsData)

	if len(wallets) == 0 {
		updateStatus("No EVM wallets configured")
		return nil
	}

	updateStatus(fmt.Sprintf("Found %d EVM wallet(s) to sync", len(wallets)))

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Initializing CoinStats client...")
	c, err := coinstats.InitiateClient()
	if err != nil {
		return err
	}

	for i, wallet := range wallets {
		updateStatus(fmt.Sprintf("Processing wallet %d/%d: %s", i+1, len(wallets), wallet.Label))
		updateStatus(fmt.Sprintf("[%s] Building request...", wallet.Label))
		req, err := c.BuildEVMBalancesRequest(ctx, wallet.Address)
		if err != nil {
			return fmt.Errorf("build request for %s: %w", wallet.Label, err)
		}

		updateStatus(fmt.Sprintf("[%s] Fetching balances from CoinStats API...", wallet.Label))
		resp := misc.DoWithRetry(ctx, req)
		if resp.Err != nil {
			return resp.Err
		}

		var positions []grist.Upsert

		// Coinstats API sometimes replies with duplicates in the payload, we have to remove these for grist
		exists := make(map[string]bool)

		var evmMultiChainBalances []coinstats.EvmMultiChainBalances
		if err := json.Unmarshal(resp.Body, &evmMultiChainBalances); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}

		for _, evmChainBalances := range evmMultiChainBalances {
			chain := evmChainBalances.Blockchain
			if strings.Contains(chainsToUpsert, chain) {
				for _, balance := range evmChainBalances.Balances {
					ticker := c.NormalizeTickerForGrist(balance.Symbol)
					contractAddress := balance.ContractAddress

					key := fmt.Sprintf("%s-%s-%s", chain, ticker, contractAddress)
					if !exists[key] {
						exists[key] = true
						updateStatus(fmt.Sprintf("[%s] Adding to upsert payload: %s %s %f", wallet.Label, chain, ticker, balance.Amount))
						positions = append(positions, grist.Upsert{
							Require: map[string]any{
								"Wallet":           wallet.Label,
								"Chain":            c.NormalizeBlockchainForGrist(chain),
								"Contract_Address": contractAddress,
							},
							Fields: map[string]any{
								"Ticker":     ticker,
								"Amount":     balance.Amount,
								"Asset_Type": token.IsStableOrVolatile(ticker),
							},
						})
					}
				}
			}
		}

		updateStatus(fmt.Sprintf("[%s] Upserting %d positions to Grist...", wallet.Label, len(positions)))
		if err := g.UpsertRecords(ctx, "Positions_Crypto_", positions, grist.UpsertOpts{}); err != nil {
			return err
		}

		updateStatus(fmt.Sprintf("[%s] ✓ Synced successfully", wallet.Label))
	}

	updateStatus(fmt.Sprintf("✓ All %d EVM wallets synced successfully", len(wallets)))
	return nil
}
