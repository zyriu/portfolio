package update_kraken

import (
	"context"
	"fmt"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/kraken"
)

func updateTrades(ctx context.Context, k kraken.Kraken, g grist.Grist, statusChan chan<- string) error {
	statusChan <- "Fetching aggregated trade count from Grist..."
	count, err := g.GetAggregatedTradesCount(ctx, "Kraken")
	if err != nil {
		return err
	}

	statusChan <- "Fetching total trade count from Kraken..."
	t, err := k.GetTradesHistory(ctx, 0)
	if err != nil {
		return err
	}

	if count == t.Result.Count {
		statusChan <- "No new trades found"
		return nil
	}

	statusChan <- "Fetching trade book from Grist..."
	book, err := g.FetchBook(ctx)
	if err != nil {
		return err
	}

	offset := t.Result.Count - count - 50
	statusChan <- fmt.Sprintf("Starting at offset %d", offset)

	rawTrades := make([]kraken.Trade, 0)
	lookup := make(map[string]bool)
	for ; offset != 0; offset -= 50 {
		if offset < 0 {
			offset = 0
		}

		t, err = k.GetTradesHistory(ctx, offset)
		if err != nil {
			if strings.Contains(err.Error(), "Rate") || strings.Contains(err.Error(), "EAPI:Rate limit exceeded") {
				statusChan <- fmt.Sprintf("⚠️ Rate limit hit, processing %d trades from Kraken...", len(rawTrades))
				break
			}

			return err
		}

		for id, t := range t.Result.Trades {
			if lookup[id] {
				continue
			}

			lookup[id] = true
			trade := t
			trade.TradeID = id
			rawTrades = append(rawTrades, trade)
		}
	}

	trades := processTrades(rawTrades, k)

	var upsert []grist.Upsert
	for _, trade := range trades {
		trade = bookTrade(&book, trade)
		upsert = append(upsert, g.CreateRecordFromTrade(trade))
	}

	if len(upsert) > 0 {
		statusChan <- fmt.Sprintf("Upserting %d trades to Grist...", len(upsert))
		if err := g.UpsertRecords(ctx, "Trades", upsert, grist.UpsertOpts{}); err != nil {
			return err
		}

		statusChan <- "Updating trade book..."
		book := g.CreateRecordsFromBook(book)
		if err := g.UpsertRecords(ctx, "Book", book, grist.UpsertOpts{}); err != nil {
			return err
		}
		statusChan <- fmt.Sprintf("✓ Successfully synced %d trades", len(upsert))
	}

	return nil
}
