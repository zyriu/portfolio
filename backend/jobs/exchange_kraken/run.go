package exchange_kraken

import (
	"context"
	"fmt"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/kraken"
)

func Run(ctx context.Context, _ ...any) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus("Initializing Kraken client...")
	k, err := kraken.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Initializing Grist client...")
	g, err := grist.InitiateClient()
	if err != nil {
		return err
	}

	updateStatus("Fetching balances...")
	if err := updateBalances(ctx, k, g); err != nil {
		return fmt.Errorf("update balances: %w", err)
	}
	updateStatus("Balances updated")

	updateStatus("Fetching trades...")
	if err := updateTrades(ctx, k, g); err != nil {
		return fmt.Errorf("update trades: %w", err)
	}
	updateStatus("Trades updated")

	updateStatus("Kraken sync completed")
	return nil
}
