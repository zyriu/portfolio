package update_kraken

import (
	"context"
	"fmt"
	"sync"

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

	// Execute updates concurrently
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

	// Update balances concurrently
	go func() {
		defer wg.Done()
		statusChan <- "Fetching balances..."
		if err := updateBalances(ctx, k, g); err != nil {
			errChan <- fmt.Errorf("update balances: %w", err)
			return
		}
		statusChan <- "Balances updated"
	}()

	// Update trades concurrently
	go func() {
		defer wg.Done()
		statusChan <- "Fetching trades..."
		if err := updateTrades(ctx, k, g, statusChan); err != nil {
			errChan <- fmt.Errorf("update trades: %w", err)
			return
		}
		statusChan <- "Trades updated"
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

	updateStatus("Kraken sync completed")
	return nil
}
