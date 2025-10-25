package update_kraken

import (
	"fmt"
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
}
