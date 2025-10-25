package update_hyperliquid

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/hyperliquid"
	"github.com/zyriu/portfolio/backend/helpers/jobstatus"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/token"
	"github.com/zyriu/portfolio/backend/helpers/trades"
)

func bookTrade(book *grist.Book, trade grist.Trade) grist.Trade {
	key := fmt.Sprintf("Hyperliquid-%s-%s", trade.Market, trade.Ticker)

	entry := grist.BookEntry{}
	if e, ok := (*book)[key]; ok {
		entry = e
	} else {
		entry.Exchange = trade.Exchange
		entry.AssetType = "Token"
		entry.Ticker = trade.Ticker
		entry.Market = trade.Market
	}

	trade, entry = trades.UpdateBookEntry(trade, entry)
	(*book)[key] = entry

	return trade
}

func processFills(ctx context.Context, h hyperliquid.Hyperliquid, g grist.Grist, userFills []hyperliquid.UserFill) ([]grist.Trade, error) {
	tradesSlice := make([]grist.Trade, 0, len(userFills))

	for _, f := range userFills {
		price, err := strconv.ParseFloat(f.Px, 64)
		if err != nil {
			return tradesSlice, err
		}

		size, err := strconv.ParseFloat(f.Sz, 64)
		if err != nil {
			return tradesSlice, err
		}

		direction := "Buy"
		if strings.EqualFold(f.Side, "A") {
			direction = "Sell"
		}

		market := "Spot"
		if h.IsPerpetuals(f.Coin) {
			market = "Futures"
		}

		base := h.NormalizeTicker(f.Coin)
		feeCurrency := h.NormalizeTicker(f.FeeToken)

		fee, err := strconv.ParseFloat(f.Fee, 64)
		if err != nil {
			return tradesSlice, err
		}

		var feeUSD float64
		if token.IsStablecoin(feeCurrency) {
			feeUSD = fee
		} else {
			feeUSD = 0 // TODO get historical price and compute fee usd
		}

		tradesSlice = append(tradesSlice, grist.Trade{
			Time:        f.Time,
			OrderValue:  price * size,
			Direction:   direction,
			Exchange:    "Hyperliquid",
			OrderType:   "Market",
			Market:      market,
			Price:       price,
			OrderSize:   size,
			Ticker:      base,
			Fee:         fee,
			FeeCurrency: feeCurrency,
			FeeUSD:      feeUSD,
			PnL:         0,
			TradeID:     strconv.FormatInt(f.Tid, 10),
		})
	}

	return tradesSlice, nil
}

func updateTrades(ctx context.Context, h hyperliquid.Hyperliquid, g grist.Grist, wallet misc.Wallet) error {
	updateStatus := jobstatus.GetStatusUpdater(ctx)

	updateStatus(fmt.Sprintf("[%s] Checking for latest trades in Grist...", wallet.Label))
	latestTrades, err := g.GetExchangeLatestTrades(ctx, "Hyperliquid", 1)
	if err != nil {
		return err
	}

	updateStatus(fmt.Sprintf("[%s] Loading trade book...", wallet.Label))
	book, err := g.FetchBook(ctx)
	if err != nil {
		return err
	}

	var seed int64
	if len(latestTrades) > 0 {
		// Note: Add 1ms to prevent inclusion of already processed latest trade
		seed = latestTrades[0].Time + 1
		updateStatus(fmt.Sprintf("[%s] Fetching trades after timestamp %d...", wallet.Label, seed))
	} else {
		seed = 0
		updateStatus(fmt.Sprintf("[%s] No existing trades found, fetching all trades...", wallet.Label))
	}

	var trades []grist.Upsert
	totalFills := 0

	for {
		var fills []hyperliquid.UserFill
		if seed == 0 {
			updateStatus(fmt.Sprintf("[%s] Fetching all user fills from Hyperliquid...", wallet.Label))
			fills, err = h.GetUserFills(ctx, wallet.Address)
			if err != nil {
				return err
			}

			// edge case to fix missing first entry from hyperliquid api and get lifetime accurate records
			for i, f := range fills {
				if f.Tid == 814936528749872 {
					fills[i].Sz = "10900"
					break
				}
			}
		} else {
			updateStatus(fmt.Sprintf("[%s] Fetching user fills from timestamp %d...", wallet.Label, seed))
			fills, err = h.GetUserFillsByTime(ctx, wallet.Address, seed)
			if err != nil {
				return err
			}
		}

		if len(fills) == 0 {
			updateStatus(fmt.Sprintf("[%s] No new fills found", wallet.Label))
			return nil
		}

		updateStatus(fmt.Sprintf("[%s] Processing %d fills...", wallet.Label, len(fills)))
		totalFills += len(fills)

		tradesSlice, err := processFills(ctx, h, g, fills)
		if err != nil {
			return fmt.Errorf("generate upserts: %w", err)
		}

		updateStatus(fmt.Sprintf("[%s] Sorting trades...", wallet.Label))
		sort.SliceStable(tradesSlice, func(i, j int) bool {
			if tradesSlice[i].Time == tradesSlice[j].Time {
				return tradesSlice[i].TradeID < tradesSlice[j].TradeID
			}

			return tradesSlice[i].Time < tradesSlice[j].Time
		})

		updateStatus(fmt.Sprintf("[%s] Booking %d trades...", wallet.Label, len(tradesSlice)))
		for _, trade := range tradesSlice {
			trade = bookTrade(&book, trade)
			trades = append(trades, g.CreateRecordFromTrade(trade))
		}

		if len(fills) < 2000 {
			break
		}

		seed = fills[2000-1].Time + 1 // Move seed up by 1ms to avoid refetching the last entry
		updateStatus(fmt.Sprintf("[%s] More fills available, continuing with next batch...", wallet.Label))
	}

	if len(trades) > 0 {
		updateStatus(fmt.Sprintf("[%s] Upserting %d trades to Grist...", wallet.Label, len(trades)))
		if err := g.UpsertRecords(ctx, "Trades", trades, grist.UpsertOpts{}); err != nil {
			return err
		}

		updateStatus(fmt.Sprintf("[%s] Updating trade book...", wallet.Label))
		book := g.CreateRecordsFromBook(book)
		if err := g.UpsertRecords(ctx, "Book", book, grist.UpsertOpts{}); err != nil {
			return err
		}
		updateStatus(fmt.Sprintf("[%s] ✓ Successfully synced %d trades from %d fills", wallet.Label, len(trades), totalFills))
	} else {
		updateStatus(fmt.Sprintf("[%s] No new trades to sync", wallet.Label))
	}

	return nil
}
