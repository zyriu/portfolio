package update_hyperliquid

import (
	"context"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/hyperliquid"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/token"
)

func updateBalances(ctx context.Context, h hyperliquid.Hyperliquid, g grist.Grist, wallet misc.Wallet) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus(fmt.Sprintf("[%s] Fetching spot balances from Hyperliquid API...", wallet.Label))
	balances, err := h.GetSpotBalances(context.Background(), wallet.Address)
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("[%s] Found %d balance(s)", wallet.Label, len(balances.Balances)))
	var upserts []grist.Upsert
	for _, balance := range balances.Balances {
		ticker := h.NormalizeTicker(balance.Coin)
		updateStatus(fmt.Sprintf("[%s] Processing balance: %s (%s)", wallet.Label, ticker, balance.Total))
		upserts = append(upserts, grist.Upsert{
			Require: map[string]any{
				"Ticker": ticker,
				"Wallet": wallet.Label,
				"Chain":  "Hyperliquid",
			},
			Fields: map[string]any{
				"Amount":     balance.Total,
				"Asset_Type": token.IsStableOrVolatile(ticker),
			},
		})
	}

	updateStatus(fmt.Sprintf("[%s] Upserting %d balance(s) to Grist...", wallet.Label, len(upserts)))
	if err := g.UpsertRecords(ctx, "Positions_Crypto_", upserts, grist.UpsertOpts{}); err != nil {
		return err
	}

	return nil
}
