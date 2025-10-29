package exchange_kraken

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
		feeUSD := fee
		if !token.IsStablecoin(quote) {
			feeUSD *= price
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
