package exchange_kraken

import (
	"testing"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/kraken"
)

// Helper to create a test Kraken instance
func createTestKraken() kraken.Kraken {
	return kraken.Kraken{
		ApiKey:    "test-key",
		ApiSecret: "test-secret",
	}
}

func TestProcessTrades_SingleTrade(t *testing.T) {
	k := createTestKraken()

	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	trade := processed[0]
	if trade.Ticker != "BTC" {
		t.Errorf("expected ticker BTC, got %s", trade.Ticker)
	}
	if trade.Exchange != "Kraken" {
		t.Errorf("expected exchange Kraken, got %s", trade.Exchange)
	}
	if trade.Direction != "Buy" {
		t.Errorf("expected direction Buy, got %s", trade.Direction)
	}
	if trade.Market != "Spot" {
		t.Errorf("expected market Spot, got %s", trade.Market)
	}
	if trade.OrderType != "Market" {
		t.Errorf("expected order type Market, got %s", trade.OrderType)
	}
	if trade.Price != 50000.0 {
		t.Errorf("expected price 50000.0, got %f", trade.Price)
	}
	if trade.OrderSize != 1.0 {
		t.Errorf("expected order size 1.0, got %f", trade.OrderSize)
	}
	if trade.OrderValue != 50000.0 {
		t.Errorf("expected order value 50000.0, got %f", trade.OrderValue)
	}
	if trade.Fee != 0.1 {
		t.Errorf("expected fee 0.1, got %f", trade.Fee)
	}
	if trade.AggregatedTrades != 1 {
		t.Errorf("expected aggregated trades 1, got %d", trade.AggregatedTrades)
	}
}

func TestProcessTrades_AggregateSameMinute(t *testing.T) {
	k := createTestKraken()

	// Two trades in the same minute with same price and type
	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0, // Same minute
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "0.5",
			Cost:      "25000.0",
			Fee:       "0.05",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade2",
			Pair:      "XXBTZUSD",
			Time:      1640000020.0, // Same minute (within 60 seconds)
			Type:      "buy",
			Price:     "50000.0", // Same price
			Vol:       "0.5",
			Cost:      "25000.0",
			Fee:       "0.05",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 1 {
		t.Fatalf("expected 1 aggregated trade, got %d", len(processed))
	}

	trade := processed[0]
	if trade.OrderSize != 1.0 {
		t.Errorf("expected aggregated size 1.0, got %f", trade.OrderSize)
	}
	if trade.OrderValue != 50000.0 {
		t.Errorf("expected aggregated value 50000.0, got %f", trade.OrderValue)
	}
	if trade.Fee != 0.1 {
		t.Errorf("expected aggregated fee 0.1, got %f", trade.Fee)
	}
	if trade.AggregatedTrades != 2 {
		t.Errorf("expected aggregated trades count 2, got %d", trade.AggregatedTrades)
	}
	// Should use earliest trade ID and time
	if trade.TradeID != "trade1" {
		t.Errorf("expected earliest trade ID trade1, got %s", trade.TradeID)
	}
}

func TestProcessTrades_DifferentPrices(t *testing.T) {
	k := createTestKraken()

	// Trades with different prices should not aggregate
	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade2",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0, // Same minute
			Type:      "buy",
			Price:     "50001.0", // Different price
			Vol:       "1.0",
			Cost:      "50001.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 2 {
		t.Fatalf("expected 2 separate trades (different prices), got %d", len(processed))
	}
}

func TestProcessTrades_DifferentTypes(t *testing.T) {
	k := createTestKraken()

	// Buy and sell should not aggregate
	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade2",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0, // Same minute, same price
			Type:      "sell",       // Different type
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 2 {
		t.Fatalf("expected 2 separate trades (buy vs sell), got %d", len(processed))
	}

	if processed[0].Direction != "Buy" {
		t.Errorf("first trade should be Buy, got %s", processed[0].Direction)
	}
	if processed[1].Direction != "Sell" {
		t.Errorf("second trade should be Sell, got %s", processed[1].Direction)
	}
}

func TestProcessTrades_MakerOrderType(t *testing.T) {
	k := createTestKraken()

	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     true, // Maker order
			OrderType: "limit",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	if processed[0].OrderType != "Limit" {
		t.Errorf("expected order type Limit for maker, got %s", processed[0].OrderType)
	}
}

func TestProcessTrades_Sorting(t *testing.T) {
	k := createTestKraken()

	// Trades in reverse chronological order
	trades := []kraken.Trade{
		{
			TradeID:   "trade3",
			Pair:      "XXBTZUSD",
			Time:      1640000120.0, // Latest
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0, // Earliest
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade2",
			Pair:      "XXBTZUSD",
			Time:      1640000060.0, // Middle
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 3 {
		t.Fatalf("expected 3 processed trades, got %d", len(processed))
	}

	// Check chronological order
	if processed[0].Time >= processed[1].Time {
		t.Errorf("first trade should be before second trade")
	}
	if processed[1].Time >= processed[2].Time {
		t.Errorf("second trade should be before third trade")
	}

	// Check trade IDs
	if processed[0].TradeID != "trade1" {
		t.Errorf("expected first trade ID trade1, got %s", processed[0].TradeID)
	}
	if processed[2].TradeID != "trade3" {
		t.Errorf("expected last trade ID trade3, got %s", processed[2].TradeID)
	}
}

func TestProcessTrades_SameTimeDifferentTradeIDs(t *testing.T) {
	k := createTestKraken()

	// Trades with same time, aggregation uses earliest time
	trades := []kraken.Trade{
		{
			TradeID:   "trade2",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000001.0, // Slightly later time, but same minute
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 1 { // Should aggregate since same price, type, and minute
		t.Fatalf("expected 1 aggregated trade, got %d", len(processed))
	}

	// Should use earliest trade ID (by time)
	if processed[0].TradeID != "trade2" {
		t.Errorf("expected earliest trade ID trade2 (by time), got %s", processed[0].TradeID)
	}
}

func TestProcessTrades_MultiplePairs(t *testing.T) {
	k := createTestKraken()

	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
		{
			TradeID:   "trade2",
			Pair:      "XETHZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "3000.0",
			Vol:       "10.0",
			Cost:      "30000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 2 {
		t.Fatalf("expected 2 processed trades (different pairs), got %d", len(processed))
	}

	// Find BTC and ETH trades
	var btcTrade, ethTrade *grist.Trade
	for i := range processed {
		if processed[i].Ticker == "BTC" {
			btcTrade = &processed[i]
		} else if processed[i].Ticker == "ETH" {
			ethTrade = &processed[i]
		}
	}

	if btcTrade == nil {
		t.Fatal("expected BTC trade not found")
	}
	if ethTrade == nil {
		t.Fatal("expected ETH trade not found")
	}

	if btcTrade.Price != 50000.0 {
		t.Errorf("expected BTC price 50000.0, got %f", btcTrade.Price)
	}
	if ethTrade.Price != 3000.0 {
		t.Errorf("expected ETH price 3000.0, got %f", ethTrade.Price)
	}
}

func TestBookTrade_NewEntry(t *testing.T) {
	book := make(grist.Book)
	trade := grist.Trade{
		Exchange:  "Kraken",
		Ticker:    "BTC",
		Market:    "Spot",
		Direction: "Buy",
		Price:     50000.0,
		OrderSize: 1.0,
	}

	result := bookTrade(&book, trade)

	key := "Kraken-Spot-BTC"
	entry, exists := book[key]
	if !exists {
		t.Fatal("expected book entry to be created")
	}

	if entry.Exchange != "Kraken" {
		t.Errorf("expected exchange Kraken, got %s", entry.Exchange)
	}
	if entry.Ticker != "BTC" {
		t.Errorf("expected ticker BTC, got %s", entry.Ticker)
	}
	if entry.Market != "Spot" {
		t.Errorf("expected market Spot, got %s", entry.Market)
	}
	if entry.AssetType != "Token" {
		t.Errorf("expected asset type Token, got %s", entry.AssetType)
	}

	// Check that trade was updated with PnL and OrderValue
	if result.PnL == 0 && result.OrderValue != 50000.0 {
		t.Errorf("expected order value 50000.0, got %f", result.OrderValue)
	}
}

func TestBookTrade_ExistingEntry(t *testing.T) {
	book := make(grist.Book)
	key := "Kraken-Spot-BTC"

	// Create existing entry
	book[key] = grist.BookEntry{
		Exchange:     "Kraken",
		Ticker:       "BTC",
		Market:       "Spot",
		AssetType:    "Token",
		PositionSize: 5.0,
		AveragePrice: 40000.0,
		CostBasis:    200000.0,
	}

	trade := grist.Trade{
		Exchange:  "Kraken",
		Ticker:    "BTC",
		Market:    "Spot",
		Direction: "Buy",
		Price:     50000.0,
		OrderSize: 2.0,
	}

	result := bookTrade(&book, trade)

	entry := book[key]
	// After buying 2 more at 50000, average should be updated
	expectedPos := 7.0
	expectedCostBasis := 200000.0 + 100000.0 // 300000
	expectedAvg := expectedCostBasis / expectedPos

	if entry.PositionSize != expectedPos {
		t.Errorf("expected position size %f, got %f", expectedPos, entry.PositionSize)
	}
	if entry.AveragePrice != expectedAvg {
		t.Errorf("expected average price %f, got %f", expectedAvg, entry.AveragePrice)
	}
	if entry.CostBasis != expectedCostBasis {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, entry.CostBasis)
	}

	// Trade should have order value set
	if result.OrderValue != 100000.0 {
		t.Errorf("expected order value 100000.0, got %f", result.OrderValue)
	}
}

func TestBookTrade_SpotSell(t *testing.T) {
	book := make(grist.Book)
	key := "Kraken-Spot-BTC"

	// Existing long position
	book[key] = grist.BookEntry{
		Exchange:     "Kraken",
		Ticker:       "BTC",
		Market:       "Spot",
		AssetType:    "Token",
		PositionSize: 10.0,
		AveragePrice: 40000.0,
		CostBasis:    400000.0,
	}

	trade := grist.Trade{
		Exchange:  "Kraken",
		Ticker:    "BTC",
		Market:    "Spot",
		Direction: "Sell",
		Price:     50000.0,
		OrderSize: 3.0,
	}

	result := bookTrade(&book, trade)

	entry := book[key]
	expectedPos := 7.0
	expectedPnL := (50000.0 - 40000.0) * 3.0 // 30000

	if entry.PositionSize != expectedPos {
		t.Errorf("expected position size %f, got %f", expectedPos, entry.PositionSize)
	}
	if result.PnL != expectedPnL {
		t.Errorf("expected PnL %f, got %f", expectedPnL, result.PnL)
	}
	if entry.AveragePrice != 40000.0 {
		t.Errorf("expected average price to remain 40000.0, got %f", entry.AveragePrice)
	}
}

func TestBookTrade_FuturesMarket(t *testing.T) {
	book := make(grist.Book)
	trade := grist.Trade{
		Exchange:  "Kraken",
		Ticker:    "BTC",
		Market:    "Futures",
		Direction: "Buy",
		Price:     50000.0,
		OrderSize: 1.0,
	}

	result := bookTrade(&book, trade)

	key := "Kraken-Futures-BTC"
	entry, exists := book[key]
	if !exists {
		t.Fatal("expected book entry to be created")
	}

	if entry.Market != "Futures" {
		t.Errorf("expected market Futures, got %s", entry.Market)
	}

	// Should use futures logic from trades.UpdateBookEntry
	if entry.PositionSize != 1.0 {
		t.Errorf("expected position size 1.0, got %f", entry.PositionSize)
	}

	if result.OrderValue != 50000.0 {
		t.Errorf("expected order value 50000.0, got %f", result.OrderValue)
	}
}

func TestProcessTrades_EdgeCase_EmptyTrades(t *testing.T) {
	k := createTestKraken()

	trades := []kraken.Trade{}
	processed := processTrades(trades, k)

	if len(processed) != 0 {
		t.Errorf("expected 0 processed trades, got %d", len(processed))
	}
}

func TestProcessTrades_EdgeCase_InvalidPrice(t *testing.T) {
	k := createTestKraken()

	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "invalid", // Invalid price
			Vol:       "1.0",
			Cost:      "50000.0",
			Fee:       "0.1",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	// Price should be 0 for invalid parse
	if processed[0].Price != 0.0 {
		t.Errorf("expected price 0.0 for invalid parse, got %f", processed[0].Price)
	}
}

func TestProcessTrades_EdgeCase_VerySmallVolumes(t *testing.T) {
	k := createTestKraken()

	trades := []kraken.Trade{
		{
			TradeID:   "trade1",
			Pair:      "XXBTZUSD",
			Time:      1640000000.0,
			Type:      "buy",
			Price:     "50000.0",
			Vol:       "0.00000001", // Very small
			Cost:      "0.0005",
			Fee:       "0.0000000001",
			Maker:     false,
			OrderType: "market",
		},
	}

	processed := processTrades(trades, k)

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	// Should handle very small volumes
	if processed[0].OrderSize <= 0 {
		t.Errorf("expected positive order size, got %f", processed[0].OrderSize)
	}
}
