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

func generateConfigForAggregation() trades.TradeConfig {
	return trades.TradeConfig{
		GetAsset: func(t any) string {
			return t.(*kraken.Trade).Pair
		},
		GetTimeMin: func(t any) int64 {
			return int64(t.(*kraken.Trade).Time) / 60
		},
		GetDirection: func(t any) string {
			return t.(*kraken.Trade).Type
		},
		GetPrice: func(t any) string {
			return t.(*kraken.Trade).Price
		},
		GetSize: func(t any) string {
			return t.(*kraken.Trade).Vol
		},
		GetFee: func(t any) string {
			return t.(*kraken.Trade).Fee
		},
		GetCost: func(t any) string {
			return t.(*kraken.Trade).Cost
		},
		GetTime: func(t any) any {
			return t.(*kraken.Trade).Time
		},
		GetTradeID: func(t any) any {
			return t.(*kraken.Trade).TradeID
		},
		UpdateSize: func(t any, v string) {
			t.(*kraken.Trade).Vol = v
		},
		UpdateFee: func(t any, v string) {
			t.(*kraken.Trade).Fee = v
		},
		UpdateCost: func(t any, v string) {
			t.(*kraken.Trade).Cost = v
		},
		UpdateTime: func(t any, v any) {
			if time, ok := v.(float64); ok {
				t.(*kraken.Trade).Time = time
			}
		},
		UpdateTradeID: func(t any, v any) {
			t.(*kraken.Trade).TradeID = v
		},
	}
}

func processTrades(tradesList []kraken.Trade, k kraken.Kraken) []grist.Trade {
	tradesInterface := make([]any, len(tradesList))
	for i := range tradesList {
		tradesInterface[i] = &tradesList[i]
	}

	config := generateConfigForAggregation()
	aggMap := trades.AggregateTrades(tradesInterface, config)

	processedTrades := make([]grist.Trade, 0, len(aggMap))
	for _, entry := range aggMap {
		trade := entry.Trade.(*kraken.Trade)
		base, quote := k.GetBaseAndQuote(trade.Pair)

		price, _ := strconv.ParseFloat(trade.Price, 64)
		size, _ := strconv.ParseFloat(trade.Vol, 64)

		fee, _ := strconv.ParseFloat(trade.Fee, 64)
		feeUSD := fee
		if !token.IsStablecoin(quote) {
			feeUSD *= price
		}

		orderType := "Market"
		if trade.Maker {
			orderType = "Limit"
		}

		processedTrades = append(processedTrades, grist.Trade{
			Ticker:           base,
			TradeID:          trade.TradeID.(string),
			Time:             int64(trade.Time * 1000),
			Exchange:         "Kraken",
			Direction:        misc.Capitalize(trade.Type),
			Fee:              fee,
			FeeCurrency:      quote,
			FeeUSD:           feeUSD,
			OrderValue:       price * size,
			OrderType:        orderType,
			OrderSize:        size,
			Market:           "Spot",
			Price:            price,
			AggregatedTrades: entry.Count,
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
