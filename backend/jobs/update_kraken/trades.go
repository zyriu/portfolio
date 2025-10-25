package update_kraken

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/kraken"
)

func updateTrades(ctx context.Context, k kraken.Kraken, g grist.Grist, statusChan chan<- string) error {
	statusChan <- "Checking existing trades in Grist..."
	count, err := g.GetMatchingRecordsAmount(ctx, "Trades", "\"Exchange\" = 'Kraken'")
	if err != nil {
		return err
	}

	statusChan <- "Loading trade book..."
	book, err := g.FetchBook(ctx)
	if err != nil {
		return err
	}

	var trades []grist.Upsert

	statusChan <- "Fetching total trade count from Kraken..."
	start, err := k.GetTradesHistory(ctx, 0)
	if err != nil {
		return err
	}

	totalTrades := start.Result.Count
	tradesToFetch := totalTrades - count
	statusChan <- fmt.Sprintf("Found %d trades (%d new to fetch)", totalTrades, tradesToFetch)

	if tradesToFetch <= 0 {
		statusChan <- "No new trades to fetch"
		return nil
	}

	batchesTotal := ((tradesToFetch - 1) / 50) + 1
	currentBatch := 0

	for ofs := ((start.Result.Count - 1 - count) / 50) * 50; ofs >= 0; ofs -= 50 {
		currentBatch++
		statusChan <- fmt.Sprintf("Fetching batch %d/%d (offset: %d)", currentBatch, batchesTotal, ofs)
		var tradesHistory kraken.TradesHistory
		retryCount := 0
		for {
			tradesHistory, err = k.GetTradesHistory(ctx, ofs)
			if err != nil {
				if strings.Contains(err.Error(), "Rate") || strings.Contains(err.Error(), "EAPI:Rate limit exceeded") {
					retryCount++
					backoff := 20 * time.Second
					statusChan <- fmt.Sprintf("⚠️ Rate limit hit (attempt %d) - waiting %ds...", retryCount, int(backoff.Seconds()))
					time.Sleep(backoff)
					continue
				}

				return err
			}

			break
		}

		if retryCount > 0 {
			statusChan <- fmt.Sprintf("✓ Retry successful, continuing batch %d/%d", currentBatch, batchesTotal)
		}

		count := tradesHistory.Result.Count
		if count == 0 {
			return nil
		}

		tradesSlice, err := processTradesHistory(tradesHistory, k)
		if err != nil {
			return err
		}

		sort.SliceStable(tradesSlice, func(i, j int) bool {
			if tradesSlice[i].Time == tradesSlice[j].Time {
				return tradesSlice[i].TradeID < tradesSlice[j].TradeID
			}

			return tradesSlice[i].Time < tradesSlice[j].Time
		})

		for _, trade := range tradesSlice {
			trade = bookTrade(&book, trade)
			trades = append(trades, g.CreateRecordFromTrade(trade))
		}
	}

	if len(trades) > 0 {
		statusChan <- fmt.Sprintf("Upserting %d trades to Grist...", len(trades))
		if err := g.UpsertRecords(ctx, "Trades", trades, grist.UpsertOpts{}); err != nil {
			return err
		}

		statusChan <- "Updating trade book..."
		book := g.CreateRecordsFromBook(book)
		if err := g.UpsertRecords(ctx, "Book", book, grist.UpsertOpts{}); err != nil {
			return err
		}
		statusChan <- fmt.Sprintf("✓ Successfully synced %d trades", len(trades))
	}

	return nil
}
