package grist

import (
	"time"

	"github.com/zyriu/portfolio/backend/helpers/misc"
)

func (g *Grist) CreateRecordsFromBook(entries Book) []Upsert {
	var records []Upsert
	for _, entry := range entries {
		records = append(records, Upsert{
			Require: map[string]any{
				"Exchange": entry.Exchange,
				"Ticker":   entry.Ticker,
				"Market":   entry.Market,
			},
			Fields: map[string]any{
				"Asset_Type":    entry.AssetType,
				"Average_Price": entry.AveragePrice,
				"Position_Size": entry.PositionSize,
				"Cost_Basis":    entry.CostBasis,
			},
		})

	}

	return records
}

func (g *Grist) CreateRecordFromTrade(trade Trade) Upsert {
	return Upsert{
		Require: map[string]any{
			"Trade_ID": trade.TradeID,
			"Exchange": trade.Exchange,
		},
		Fields: map[string]any{
			"Ticker":            trade.Ticker,
			"PnL":               trade.PnL,
			"Order_Value":       trade.OrderValue,
			"Time":              trade.Time,
			"Date":              time.UnixMilli(trade.Time).UTC().Format("02 Jan 2006 0304 pm"),
			"Direction":         trade.Direction,
			"Market":            trade.Market,
			"Fee":               trade.Fee,
			"Fee_Currency":      trade.FeeCurrency,
			"Fee_USD_":          trade.FeeUSD,
			"Order_Type":        misc.Capitalize(trade.OrderType),
			"Price":             trade.Price,
			"Order_Size":        trade.OrderSize,
			"Aggregated_Trades": trade.AggregatedTrades,
		},
	}
}
