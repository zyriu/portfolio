package update_kraken

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/kraken"
	"github.com/zyriu/portfolio/backend/helpers/misc"
	"github.com/zyriu/portfolio/backend/helpers/token"
	"github.com/zyriu/portfolio/backend/helpers/trades"
)

func bookTrade(book *grist.Book, trade grist.Trade) grist.Trade {
	key := fmt.Sprintf("Kraken-%s-%s", trade.Market, trade.Ticker)

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

func processTrades(trades []kraken.Trade, k kraken.Kraken) []grist.Trade {
	processedTrades := make([]grist.Trade, 0)

	type aggKey struct {
		Pair    string
		TimeMin int64
		Type    string
		Price   string
	}
	type aggEntry struct {
		trade *kraken.Trade
		count int
	}
	aggMap := make(map[aggKey]*aggEntry)

	for _, trade := range trades {
		timeMin := int64(trade.Time) / 60
		key := aggKey{
			Pair:    trade.Pair,
			TimeMin: timeMin,
			Type:    trade.Type,
			Price:   trade.Price,
		}
		entry, exists := aggMap[key]
		if !exists {
			cp := trade
			aggMap[key] = &aggEntry{
				trade: &cp,
				count: 1,
			}
			continue
		}

		parseF := func(s string) float64 {
			val, _ := strconv.ParseFloat(s, 64)
			return val
		}
		formatF := func(f float64) string {
			return fmt.Sprintf("%.12g", f)
		}

		entry.trade.Vol = formatF(parseF(entry.trade.Vol) + parseF(trade.Vol))
		entry.trade.Cost = formatF(parseF(entry.trade.Cost) + parseF(trade.Cost))
		entry.trade.Fee = formatF(parseF(entry.trade.Fee) + parseF(trade.Fee))
		entry.count++

		if trade.Time < entry.trade.Time {
			entry.trade.TradeID = trade.TradeID
			entry.trade.Time = trade.Time
		}
	}

	processedTrades = processedTrades[:0]
	for _, entry := range aggMap {
		t := entry.trade
		base, quote := k.GetBaseAndQuote(t.Pair)

		price, _ := strconv.ParseFloat(t.Price, 64)
		size, _ := strconv.ParseFloat(t.Vol, 64)

		fee, _ := strconv.ParseFloat(t.Fee, 64)
		var feeUSD float64
		if token.IsStablecoin(quote) {
			feeUSD = fee
		} else {
			feeUSD = fee * price
		}

		orderType := "Market"
		if t.Maker {
			orderType = "Limit"
		}

		processedTrades = append(processedTrades, grist.Trade{
			Ticker:           base,
			TradeID:          t.TradeID.(string),
			Time:             int64(t.Time * 1000),
			Exchange:         "Kraken",
			Direction:        misc.Capitalize(t.Type),
			Fee:              fee,
			FeeCurrency:      quote,
			FeeUSD:           feeUSD,
			OrderValue:       price * size,
			OrderType:        orderType,
			OrderSize:        size,
			Market:           "Spot",
			Price:            price,
			AggregatedTrades: entry.count,
		})
	}

	sort.Slice(processedTrades, func(i, j int) bool {
		if processedTrades[i].Time == processedTrades[j].Time {
			return processedTrades[i].TradeID < processedTrades[j].TradeID
		}

		return processedTrades[i].Time < processedTrades[j].Time
	})

	return processedTrades
}

/*
func processTradesHistory(trades kraken.TradesHistory, k kraken.Kraken) ([]grist.Trade, error) {
	tradesSlice := make([]grist.Trade, 0, len(trades.Result.Trades))

	for tradeID, trade := range trades.Result.Trades {
		price, err := strconv.ParseFloat(trade.Price, 64)
		if err != nil {
			return tradesSlice, fmt.Errorf("parse price for trade %s: %w", tradeID, err)
		}

		size, err := strconv.ParseFloat(trade.Vol, 64)
		if err != nil {
			return tradesSlice, fmt.Errorf("parse volume for trade %s: %w", tradeID, err)
		}

		direction := misc.Capitalize(trade.Type)
		base, quote := k.GetBaseAndQuote(trade.Pair)
		orderType := "Market"
		if trade.Maker {
			orderType = "Limit"
		}

		fee, err := strconv.ParseFloat(trade.Fee, 64)
		if err != nil {
			return tradesSlice, fmt.Errorf("parse fee for trade %s: %w", tradeID, err)
		}

		var feeUSD float64
		if token.IsStablecoin(quote) {
			feeUSD = fee
		} else {
			feeUSD = 0 // TODO get historical price and compute fee usd
		}

		tradesSlice = append(tradesSlice, grist.Trade{
			Ticker:      base,
			TradeID:     tradeID,
			Time:        int64(trade.Time * 1000),
			Exchange:    "Kraken",
			Direction:   direction,
			Fee:         fee,
			FeeCurrency: quote,
			FeeUSD:      feeUSD,
			OrderValue:  price * size,
			OrderType:   orderType,
			OrderSize:   size,
			Market:      "Spot",
			Price:       price,
		})
	}

	return tradesSlice, nil
}*/
