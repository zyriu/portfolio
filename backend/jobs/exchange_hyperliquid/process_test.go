package exchange_hyperliquid

import (
	"testing"

	"github.com/zyriu/portfolio/backend/helpers/grist"
	"github.com/zyriu/portfolio/backend/helpers/hyperliquid"
)

// Helper to create a test Hyperliquid instance
func createTestHyperliquid() hyperliquid.Hyperliquid {
	return hyperliquid.Hyperliquid{}
}

func TestProcessFills_SingleFill(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000, // Milliseconds
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	trade := processed[0]
	if trade.Ticker != "BTC" {
		t.Errorf("expected ticker BTC, got %s", trade.Ticker)
	}
	if trade.Exchange != "Hyperliquid" {
		t.Errorf("expected exchange Hyperliquid, got %s", trade.Exchange)
	}
	if trade.Direction != "Buy" {
		t.Errorf("expected direction Buy, got %s", trade.Direction)
	}
	if trade.Market != "Futures" {
		t.Errorf("expected market Futures, got %s", trade.Market)
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
	if trade.FeeCurrency != "USDC" {
		t.Errorf("expected fee currency USDC, got %s", trade.FeeCurrency)
	}
	if trade.TradeID != "123456789" {
		t.Errorf("expected trade ID 123456789, got %s", trade.TradeID)
	}
	if trade.AggregatedTrades != 1 {
		t.Errorf("expected aggregated trades 1, got %d", trade.AggregatedTrades)
	}
}

func TestProcessFills_SellSide(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "A", // Ask = Sell
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	if processed[0].Direction != "Sell" {
		t.Errorf("expected direction Sell for side A, got %s", processed[0].Direction)
	}
}

func TestProcessFills_AggregateSameMinute(t *testing.T) {
	h := createTestHyperliquid()

	// Two fills in the same minute with same price and side
	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000, // Same minute (rounded to minutes)
			Side:      "B",
			Px:        "50000.0",
			Sz:        "0.5",
			Fee:       "0.05",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
		{
			Coin:      "BTC",
			Tid:       987654321,
			Time:      1640000020000, // Same minute (within 60 seconds when converted)
			Side:      "B",
			Px:        "50000.0", // Same price
			Sz:        "0.5",
			Fee:       "0.05",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	if trade.TradeID != "123456789" {
		t.Errorf("expected earliest trade ID 123456789, got %s", trade.TradeID)
	}
}

func TestProcessFills_DifferentPrices(t *testing.T) {
	h := createTestHyperliquid()

	// Fills with different prices should not aggregate
	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
		{
			Coin:      "BTC",
			Tid:       987654321,
			Time:      1640000000000, // Same minute
			Side:      "B",
			Px:        "50001.0", // Different price
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 2 {
		t.Fatalf("expected 2 separate trades (different prices), got %d", len(processed))
	}
}

func TestProcessFills_DifferentSides(t *testing.T) {
	h := createTestHyperliquid()

	// Buy and sell should not aggregate
	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
		{
			Coin:      "BTC",
			Tid:       987654321,
			Time:      1640000000000, // Same minute, same price
			Side:      "A",           // Different side
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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

func TestProcessFills_SpotMarket(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "@142",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	if processed[0].Market != "Spot" {
		t.Errorf("expected market Spot for @142, got %s", processed[0].Market)
	}
}

func TestProcessFills_SpecialTokens(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "UFART",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "0.001",
			Sz:        "1000.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	if processed[0].Market != "Spot" {
		t.Errorf("expected market Spot for UFART, got %s", processed[0].Market)
	}
	if processed[0].Ticker != "FARTCOIN" {
		t.Errorf("expected ticker FARTCOIN, got %s", processed[0].Ticker)
	}
}

func TestProcessFills_NonStablecoinFee(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.01", // Fee in ETH
			FeeToken:  "ETH",  // Not a stablecoin
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	// FeeUSD should be fee * price (since ETH is not stablecoin)
	// Assuming ETH price around 3000, but we don't have it, so it should be 0.01 * 50000
	// Actually, looking at the code, it uses the trade price, so feeUSD = 0.01 * 50000 = 500
	expectedFeeUSD := 0.01 * 50000.0
	if processed[0].FeeUSD != expectedFeeUSD {
		t.Errorf("expected fee USD %f, got %f", expectedFeeUSD, processed[0].FeeUSD)
	}
}

func TestProcessFills_SameTimeDifferentTradeIDs(t *testing.T) {
	h := createTestHyperliquid()

	// Fills with same minute but different times
	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       200,
			Time:      1640000005000, // Slightly later time, but same minute
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
		{
			Coin:      "BTC",
			Tid:       100,
			Time:      1640000000000, // Earlier time in same minute
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 { // Should aggregate since same price, side, and minute
		t.Fatalf("expected 1 aggregated trade, got %d", len(processed))
	}

	// Should use earliest trade ID (by time)
	if processed[0].TradeID != "100" {
		t.Errorf("expected earliest trade ID 100 (by time), got %s", processed[0].TradeID)
	}
}

func TestProcessFills_MultipleCoins(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "50000.0",
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
		{
			Coin:      "ETH",
			Tid:       987654321,
			Time:      1640000000000,
			Side:      "B",
			Px:        "3000.0",
			Sz:        "10.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 2 {
		t.Fatalf("expected 2 processed trades (different coins), got %d", len(processed))
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
		Exchange:  "Hyperliquid",
		Ticker:    "BTC",
		Market:    "Futures",
		Direction: "Buy",
		Price:     50000.0,
		OrderSize: 1.0,
	}

	result := bookTrade(&book, trade)

	key := "Hyperliquid-Futures-BTC"
	entry, exists := book[key]
	if !exists {
		t.Fatal("expected book entry to be created")
	}

	if entry.Exchange != "Hyperliquid" {
		t.Errorf("expected exchange Hyperliquid, got %s", entry.Exchange)
	}
	if entry.Ticker != "BTC" {
		t.Errorf("expected ticker BTC, got %s", entry.Ticker)
	}
	if entry.Market != "Futures" {
		t.Errorf("expected market Futures, got %s", entry.Market)
	}
	if entry.AssetType != "Token" {
		t.Errorf("expected asset type Token, got %s", entry.AssetType)
	}

	// Check that trade was updated with PnL and OrderValue
	if result.OrderValue != 50000.0 {
		t.Errorf("expected order value 50000.0, got %f", result.OrderValue)
	}
}

func TestBookTrade_ExistingEntry_FuturesLong(t *testing.T) {
	book := make(grist.Book)
	key := "Hyperliquid-Futures-BTC"

	// Create existing long position
	book[key] = grist.BookEntry{
		Exchange:     "Hyperliquid",
		Ticker:       "BTC",
		Market:       "Futures",
		AssetType:    "Token",
		PositionSize: 5.0,
		AveragePrice: 40000.0,
		CostBasis:    200000.0,
	}

	trade := grist.Trade{
		Exchange:  "Hyperliquid",
		Ticker:    "BTC",
		Market:    "Futures",
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

func TestBookTrade_FuturesShort(t *testing.T) {
	book := make(grist.Book)
	key := "Hyperliquid-Futures-BTC"

	// Existing short position
	book[key] = grist.BookEntry{
		Exchange:     "Hyperliquid",
		Ticker:       "BTC",
		Market:       "Futures",
		AssetType:    "Token",
		PositionSize: -10.0, // Short
		AveragePrice: 40000.0,
		CostBasis:    -400000.0,
	}

	trade := grist.Trade{
		Exchange:  "Hyperliquid",
		Ticker:    "BTC",
		Market:    "Futures",
		Direction: "Buy",
		Price:     35000.0, // Closing some short at profit
		OrderSize: 3.0,
	}

	result := bookTrade(&book, trade)

	entry := book[key]
	expectedPos := -7.0                      // -10 + 3
	expectedPnL := (40000.0 - 35000.0) * 3.0 // 15000 (profitable close)

	if entry.PositionSize != expectedPos {
		t.Errorf("expected position size %f, got %f", expectedPos, entry.PositionSize)
	}
	if result.PnL != expectedPnL {
		t.Errorf("expected PnL %f, got %f", expectedPnL, result.PnL)
	}
}

func TestProcessFills_EdgeCase_EmptyFills(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{}
	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 0 {
		t.Errorf("expected 0 processed trades, got %d", len(processed))
	}
}

func TestProcessFills_EdgeCase_InvalidPrice(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "invalid", // Invalid price
			Sz:        "1.0",
			Fee:       "0.1",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)

	// Should return error for invalid price
	if err == nil {
		t.Fatal("expected error for invalid price, got nil")
	}
	if len(processed) != 0 {
		t.Errorf("expected 0 processed trades on error, got %d", len(processed))
	}
}

func TestProcessFills_EdgeCase_VerySmallSizes(t *testing.T) {
	h := createTestHyperliquid()

	fills := []hyperliquid.UserFill{
		{
			Coin:      "BTC",
			Tid:       123456789,
			Time:      1640000000000,
			Side:      "B",
			Px:        "50000.0",
			Sz:        "0.00000001", // Very small
			Fee:       "0.0000000001",
			FeeToken:  "USDC",
			ClosedPnL: "0",
		},
	}

	processed, err := processFills(h, fills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(processed) != 1 {
		t.Fatalf("expected 1 processed trade, got %d", len(processed))
	}

	// Should handle very small sizes
	if processed[0].OrderSize <= 0 {
		t.Errorf("expected positive order size, got %f", processed[0].OrderSize)
	}
}
