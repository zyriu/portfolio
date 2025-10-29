package update_non_evm_balances

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/coinstats"
	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/settings"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func Run(ctx context.Context, args ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)
	blockchain := fmt.Sprint(args[0])

	updateStatus(fmt.Sprintf("Processing %s wallets...", blockchain))

	updateStatus("Loading settings...")
	settingsData, err := settings.GetCurrentSettings()
	if err != nil {
		return err
	}

	wallets := settings.GetNonEVMWallets(settingsData)

	// Filter wallets by blockchain
	var filteredWallets []settings.UnifiedWallet
	for _, wallet := range wallets {
		if wallet.Type == blockchain {
			filteredWallets = append(filteredWallets, wallet)
		}
	}

	// If no wallets configured for this blockchain, return gracefully
	if len(filteredWallets) == 0 {
		updateStatus(fmt.Sprintf("No %s wallets configured", blockchain))
		return nil
	}

	updateStatus(fmt.Sprintf("Found %d %s wallet(s) to sync", len(filteredWallets), blockchain))

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

	chain := c.NormalizeBlockchainForGrist(blockchain)

	for i, wallet := range filteredWallets {
		updateStatus(fmt.Sprintf("Processing wallet %d/%d: %s", i+1, len(filteredWallets), wallet.Label))
		updateStatus(fmt.Sprintf("[%s] Building request...", wallet.Label))
		req, err := c.BuildSingleBalanceRequest(ctx, wallet.Address, blockchain)
		if err != nil {
			return fmt.Errorf("build request for %s: %w", wallet.Label, err)
		}

		updateStatus(fmt.Sprintf("[%s] Fetching balances from CoinStats...", wallet.Label))
		resp := misc.DoWithRetry(ctx, req)
		if resp.Err != nil {
			return resp.Err
		}

		// Handle empty or null response body
		if len(resp.Body) == 0 || string(resp.Body) == "null" || string(resp.Body) == "[]" {
			updateStatus(fmt.Sprintf("[%s] No balances found", wallet.Label))
			continue
		}

		var positions []grist.Upsert
		var singleChainBalance coinstats.SingleChainBalance
		if err := json.Unmarshal(resp.Body, &singleChainBalance.Balances); err != nil {
			return fmt.Errorf("decode response for %s: %w (body: %s)", wallet.Label, err, string(resp.Body))
		}

		if len(singleChainBalance.Balances) == 0 {
			updateStatus(fmt.Sprintf("[%s] No balances found after parsing", wallet.Label))
			continue
		}

		updateStatus(fmt.Sprintf("[%s] Processing %d balance(s)...", wallet.Label, len(singleChainBalance.Balances)))
		for _, balance := range singleChainBalance.Balances {
			ticker := c.NormalizeTickerForGrist(balance.Symbol)
			contractAddress := balance.ContractAddress
			updateStatus(fmt.Sprintf("[%s]   %s: %.4f @ $%.4f", wallet.Label, ticker, balance.Amount, balance.Price))

			positions = append(positions, grist.Upsert{
				Require: map[string]any{
					"Wallet":           wallet.Label,
					"Chain":            chain,
					"Contract_Address": contractAddress,
				},
				Fields: map[string]any{
					"Ticker":     ticker,
					"Amount":     balance.Amount,
					"Asset_Type": token.GetAssetType(ticker),
				},
			})
		}

		updateStatus(fmt.Sprintf("[%s] Checking for existing records to clean up...", wallet.Label))
		query := fmt.Sprintf("filter={\"Wallet\":[\"%s\"],\"Chain\":[\"%s\"]}", wallet.Label, chain)
		matches, err := g.GetRecords(ctx, "Positions_Crypto_", query)
		if err != nil {
			return err
		}

		var recordsToDelete []int64
		for _, match := range matches.Records {
			recordsToDelete = append(recordsToDelete, match.RecordID)
		}

		if len(recordsToDelete) > 0 {
			updateStatus(fmt.Sprintf("[%s] Deleting %d old record(s)...", wallet.Label, len(recordsToDelete)))
			if err := g.DeleteRecords(ctx, "Positions_Crypto_", recordsToDelete); err != nil {
				return err
			}

			updateStatus(fmt.Sprintf("[%s] Deleted %d existing positions", wallet.Label, len(recordsToDelete)))
		} else {
			updateStatus(fmt.Sprintf("[%s] No old records to delete", wallet.Label))
		}

		updateStatus(fmt.Sprintf("[%s] Upserting %d positions to Grist...", wallet.Label, len(positions)))
		if err := g.UpsertRecords(ctx, "Positions_Crypto_", positions, grist.UpsertOpts{}); err != nil {
			return fmt.Errorf("error upserting balances: %v", err)
		}

		updateStatus(fmt.Sprintf("[%s] ✓ Synced successfully", wallet.Label))
	}

	updateStatus(fmt.Sprintf("✓ All %d %s wallets synced successfully", len(filteredWallets), blockchain))
	return nil
}
