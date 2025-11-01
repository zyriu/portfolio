package exchange_hyperliquid

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

func generateConfigForAggregation() trades.TradeConfig {
	return trades.TradeConfig{
		GetAsset: func(t any) string {
			return t.(*hyperliquid.UserFill).Coin
		},
		GetTimeMin: func(t any) int64 {
			return t.(*hyperliquid.UserFill).Time / 60000
		},
		GetDirection: func(t any) string {
			return t.(*hyperliquid.UserFill).Side
		},
		GetPrice: func(t any) string {
			return t.(*hyperliquid.UserFill).Px
		},
		GetSize: func(t any) string {
			return t.(*hyperliquid.UserFill).Sz
		},
		GetFee: func(t any) string {
			return t.(*hyperliquid.UserFill).Fee
		},
		GetCost: nil,
		GetTime: func(t any) any {
			return t.(*hyperliquid.UserFill).Time
		},
		GetTradeID: func(t any) any {
			return t.(*hyperliquid.UserFill).Tid
		},
		UpdateSize: func(t any, v string) {
			t.(*hyperliquid.UserFill).Sz = v
		},
		UpdateFee: func(t any, v string) {
			t.(*hyperliquid.UserFill).Fee = v
		},
		UpdateCost: nil,
		UpdateTime: func(t any, v any) {
			if time, ok := v.(int64); ok {
				t.(*hyperliquid.UserFill).Time = time
			}
		},
		UpdateTradeID: func(t any, v any) {
			if tid, ok := v.(int64); ok {
				t.(*hyperliquid.UserFill).Tid = tid
			}
		},
	}
}

func processFills(h hyperliquid.Hyperliquid, userFills []hyperliquid.UserFill) ([]grist.Trade, error) {
	fillsInterface := make([]any, len(userFills))
	for i := range userFills {
		fillsInterface[i] = &userFills[i]
	}

	config := generateConfigForAggregation()
	aggMap := trades.AggregateTrades(fillsInterface, config)

	tradesSlice := make([]grist.Trade, 0, len(aggMap))
	for _, entry := range aggMap {
		f := entry.Trade.(*hyperliquid.UserFill)
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
		fee, _ := strconv.ParseFloat(f.Fee, 64)
		feeUSD := fee
		if !token.IsStablecoin(feeCurrency) {
			feeUSD *= price
		}

		tradesSlice = append(tradesSlice, grist.Trade{
			Time:             f.Time,
			OrderValue:       price * size,
			Direction:        direction,
			Exchange:         "Hyperliquid",
			OrderType:        "Market",
			Market:           market,
			Price:            price,
			OrderSize:        size,
			Ticker:           base,
			Fee:              fee,
			FeeCurrency:      feeCurrency,
			FeeUSD:           feeUSD,
			PnL:              0,
			TradeID:          strconv.FormatInt(f.Tid, 10),
			AggregatedTrades: entry.Count,
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

		tradesSlice, err := processFills(h, fills)
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

		seed = fills[2000-1].Time + 1
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
