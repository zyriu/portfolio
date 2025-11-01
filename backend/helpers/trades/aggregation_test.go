package trades

import (
	"testing"
)

// Mock trade types for testing
type mockKrakenTrade struct {
	TradeID interface{}
	Pair    string
	Time    float64
	Type    string
	Price   string
	Vol     string
	Cost    string
	Fee     string
	Maker   bool
}

type mockHyperliquidFill struct {
	Tid      int64
	Coin     string
	Time     int64
	Side     string
	Px       string
	Sz       string
	Fee      string
	FeeToken string
}

func TestAggregateTrades_SingleTrade(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
	}

	config := TradeConfig{
		GetAsset: func(t interface{}) string {
			return t.(*mockKrakenTrade).Pair
		},
		GetTimeMin: func(t interface{}) int64 {
			return int64(t.(*mockKrakenTrade).Time) / 60
		},
		GetDirection: func(t interface{}) string {
			return t.(*mockKrakenTrade).Type
		},
		GetPrice: func(t interface{}) string {
			return t.(*mockKrakenTrade).Price
		},
		GetSize: func(t interface{}) string {
			return t.(*mockKrakenTrade).Vol
		},
		GetFee: func(t interface{}) string {
			return t.(*mockKrakenTrade).Fee
		},
		GetCost: func(t interface{}) string {
			return t.(*mockKrakenTrade).Cost
		},
		GetTime: func(t interface{}) interface{} {
			return t.(*mockKrakenTrade).Time
		},
		GetTradeID: func(t interface{}) interface{} {
			return t.(*mockKrakenTrade).TradeID
		},
		UpdateSize: func(t interface{}, v string) {
			t.(*mockKrakenTrade).Vol = v
		},
		UpdateFee: func(t interface{}, v string) {
			t.(*mockKrakenTrade).Fee = v
		},
		UpdateCost: func(t interface{}, v string) {
			t.(*mockKrakenTrade).Cost = v
		},
		UpdateTime: func(t interface{}, v interface{}) {
			if time, ok := v.(float64); ok {
				t.(*mockKrakenTrade).Time = time
			}
		},
		UpdateTradeID: func(t interface{}, v interface{}) {
			t.(*mockKrakenTrade).TradeID = v
		},
	}

	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 1 {
		t.Fatalf("expected 1 aggregated trade, got %d", len(aggMap))
	}

	var entry *AggregatedTrade
	for _, e := range aggMap {
		entry = e
	}

	if entry.Count != 1 {
		t.Errorf("expected count 1, got %d", entry.Count)
	}
	if entry.AggregatedSize != 1.0 {
		t.Errorf("expected aggregated size 1.0, got %f", entry.AggregatedSize)
	}
	if entry.AggregatedFee != 0.1 {
		t.Errorf("expected aggregated fee 0.1, got %f", entry.AggregatedFee)
	}
}

func TestAggregateTrades_AggregateSameMinute(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "0.5",
			Cost:    "25000.0",
			Fee:     "0.05",
		},
		&mockKrakenTrade{
			TradeID: "trade2",
			Pair:    "XXBTZUSD",
			Time:    1640000020.0, // Same minute
			Type:    "buy",
			Price:   "50000.0", // Same price
			Vol:     "0.5",
			Cost:    "25000.0",
			Fee:     "0.05",
		},
	}

	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 1 {
		t.Fatalf("expected 1 aggregated trade, got %d", len(aggMap))
	}

	var entry *AggregatedTrade
	for _, e := range aggMap {
		entry = e
	}

	if entry.Count != 2 {
		t.Errorf("expected count 2, got %d", entry.Count)
	}
	if entry.AggregatedSize != 1.0 {
		t.Errorf("expected aggregated size 1.0, got %f", entry.AggregatedSize)
	}
	if entry.AggregatedFee != 0.1 {
		t.Errorf("expected aggregated fee 0.1, got %f", entry.AggregatedFee)
	}
	if entry.AggregatedCost != 50000.0 {
		t.Errorf("expected aggregated cost 50000.0, got %f", entry.AggregatedCost)
	}
	// Should use earliest trade ID
	trade := entry.Trade.(*mockKrakenTrade)
	if trade.TradeID != "trade1" {
		t.Errorf("expected earliest trade ID trade1, got %v", trade.TradeID)
	}
}

func TestAggregateTrades_DifferentPrices(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
		&mockKrakenTrade{
			TradeID: "trade2",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50001.0", // Different price
			Vol:     "1.0",
			Cost:    "50001.0",
			Fee:     "0.1",
		},
	}

	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 2 {
		t.Fatalf("expected 2 separate trades (different prices), got %d", len(aggMap))
	}
}

func TestAggregateTrades_DifferentDirections(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
		&mockKrakenTrade{
			TradeID: "trade2",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "sell", // Different type
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
	}

	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 2 {
		t.Fatalf("expected 2 separate trades (buy vs sell), got %d", len(aggMap))
	}
}

func TestAggregateTrades_EarliestTimeAndID(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade2",
			Pair:    "XXBTZUSD",
			Time:    1640000001.0, // Later time
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0, // Earlier time
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
	}

	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 1 {
		t.Fatalf("expected 1 aggregated trade, got %d", len(aggMap))
	}

	var entry *AggregatedTrade
	for _, e := range aggMap {
		entry = e
	}

	// Should use earliest time and trade ID
	trade := entry.Trade.(*mockKrakenTrade)
	if trade.TradeID != "trade1" {
		t.Errorf("expected earliest trade ID trade1, got %v", trade.TradeID)
	}
	if trade.Time != 1640000000.0 {
		t.Errorf("expected earliest time 1640000000.0, got %f", trade.Time)
	}
}

func TestAggregateTrades_HyperliquidFormat(t *testing.T) {
	fills := []interface{}{
		&mockHyperliquidFill{
			Tid:  123456789,
			Coin: "BTC",
			Time: 1640000000000, // Milliseconds
			Side: "B",
			Px:   "50000.0",
			Sz:   "1.0",
			Fee:  "0.1",
		},
		&mockHyperliquidFill{
			Tid:  987654321,
			Coin: "BTC",
			Time: 1640000020000, // Same minute
			Side: "B",
			Px:   "50000.0",
			Sz:   "0.5",
			Fee:  "0.05",
		},
	}

	config := TradeConfig{
		GetAsset: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Coin
		},
		GetTimeMin: func(t interface{}) int64 {
			return t.(*mockHyperliquidFill).Time / 60000
		},
		GetDirection: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Side
		},
		GetPrice: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Px
		},
		GetSize: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Sz
		},
		GetFee: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Fee
		},
		GetCost: nil, // Hyperliquid doesn't have cost
		GetTime: func(t interface{}) interface{} {
			return t.(*mockHyperliquidFill).Time
		},
		GetTradeID: func(t interface{}) interface{} {
			return t.(*mockHyperliquidFill).Tid
		},
		UpdateSize: func(t interface{}, v string) {
			t.(*mockHyperliquidFill).Sz = v
		},
		UpdateFee: func(t interface{}, v string) {
			t.(*mockHyperliquidFill).Fee = v
		},
		UpdateCost: nil,
		UpdateTime: func(t interface{}, v interface{}) {
			if time, ok := v.(int64); ok {
				t.(*mockHyperliquidFill).Time = time
			}
		},
		UpdateTradeID: func(t interface{}, v interface{}) {
			if tid, ok := v.(int64); ok {
				t.(*mockHyperliquidFill).Tid = tid
			}
		},
	}

	aggMap := AggregateTrades(fills, config)

	if len(aggMap) != 1 {
		t.Fatalf("expected 1 aggregated fill, got %d", len(aggMap))
	}

	var entry *AggregatedTrade
	for _, e := range aggMap {
		entry = e
	}

	if entry.Count != 2 {
		t.Errorf("expected count 2, got %d", entry.Count)
	}
	if entry.AggregatedSize != 1.5 {
		t.Errorf("expected aggregated size 1.5, got %f", entry.AggregatedSize)
	}
	expectedFee := 0.15
	if entry.AggregatedFee < expectedFee-0.0001 || entry.AggregatedFee > expectedFee+0.0001 {
		t.Errorf("expected aggregated fee ~0.15, got %f", entry.AggregatedFee)
	}

	// Should use earliest trade ID
	fill := entry.Trade.(*mockHyperliquidFill)
	if fill.Tid != 123456789 {
		t.Errorf("expected earliest trade ID 123456789, got %d", fill.Tid)
	}
}

func TestAggregateTrades_EmptyTrades(t *testing.T) {
	trades := []interface{}{}
	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 0 {
		t.Errorf("expected 0 aggregated trades, got %d", len(aggMap))
	}
}

func TestAggregateTrades_InvalidNumericValues(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "invalid", // Invalid size
			Cost:    "50000.0",
			Fee:     "0.1",
		},
	}

	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 1 {
		t.Fatalf("expected 1 aggregated trade, got %d", len(aggMap))
	}

	var entry *AggregatedTrade
	for _, e := range aggMap {
		entry = e
	}

	// Invalid values should parse to 0
	if entry.AggregatedSize != 0.0 {
		t.Errorf("expected aggregated size 0.0 for invalid value, got %f", entry.AggregatedSize)
	}
}

func TestAggregateTrades_MultipleAssets(t *testing.T) {
	trades := []interface{}{
		&mockKrakenTrade{
			TradeID: "trade1",
			Pair:    "XXBTZUSD",
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "50000.0",
			Vol:     "1.0",
			Cost:    "50000.0",
			Fee:     "0.1",
		},
		&mockKrakenTrade{
			TradeID: "trade2",
			Pair:    "XETHZUSD", // Different asset
			Time:    1640000000.0,
			Type:    "buy",
			Price:   "3000.0",
			Vol:     "10.0",
			Cost:    "30000.0",
			Fee:     "0.1",
		},
	}

	config := createKrakenConfig()
	aggMap := AggregateTrades(trades, config)

	if len(aggMap) != 2 {
		t.Fatalf("expected 2 separate trades (different assets), got %d", len(aggMap))
	}
}

func TestAggregateTrades_NoCostField(t *testing.T) {
	// Test with Hyperliquid-style config (no cost field)
	fills := []interface{}{
		&mockHyperliquidFill{
			Tid:  123456789,
			Coin: "BTC",
			Time: 1640000000000,
			Side: "B",
			Px:   "50000.0",
			Sz:   "1.0",
			Fee:  "0.1",
		},
	}

	config := TradeConfig{
		GetAsset: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Coin
		},
		GetTimeMin: func(t interface{}) int64 {
			return t.(*mockHyperliquidFill).Time / 60000
		},
		GetDirection: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Side
		},
		GetPrice: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Px
		},
		GetSize: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Sz
		},
		GetFee: func(t interface{}) string {
			return t.(*mockHyperliquidFill).Fee
		},
		GetCost: nil, // No cost field
		GetTime: func(t interface{}) interface{} {
			return t.(*mockHyperliquidFill).Time
		},
		GetTradeID: func(t interface{}) interface{} {
			return t.(*mockHyperliquidFill).Tid
		},
		UpdateSize: func(t interface{}, v string) {
			t.(*mockHyperliquidFill).Sz = v
		},
		UpdateFee: func(t interface{}, v string) {
			t.(*mockHyperliquidFill).Fee = v
		},
		UpdateCost: nil,
		UpdateTime: func(t interface{}, v interface{}) {
			if time, ok := v.(int64); ok {
				t.(*mockHyperliquidFill).Time = time
			}
		},
		UpdateTradeID: func(t interface{}, v interface{}) {
			if tid, ok := v.(int64); ok {
				t.(*mockHyperliquidFill).Tid = tid
			}
		},
	}

	aggMap := AggregateTrades(fills, config)

	if len(aggMap) != 1 {
		t.Fatalf("expected 1 aggregated fill, got %d", len(aggMap))
	}

	var entry *AggregatedTrade
	for _, e := range aggMap {
		entry = e
	}

	// Cost should be 0 when not provided
	if entry.AggregatedCost != 0.0 {
		t.Errorf("expected aggregated cost 0.0 when no cost field, got %f", entry.AggregatedCost)
	}
}

// Helper function to create Kraken-style config
func createKrakenConfig() TradeConfig {
	return TradeConfig{
		GetAsset: func(t interface{}) string {
			return t.(*mockKrakenTrade).Pair
		},
		GetTimeMin: func(t interface{}) int64 {
			return int64(t.(*mockKrakenTrade).Time) / 60
		},
		GetDirection: func(t interface{}) string {
			return t.(*mockKrakenTrade).Type
		},
		GetPrice: func(t interface{}) string {
			return t.(*mockKrakenTrade).Price
		},
		GetSize: func(t interface{}) string {
			return t.(*mockKrakenTrade).Vol
		},
		GetFee: func(t interface{}) string {
			return t.(*mockKrakenTrade).Fee
		},
		GetCost: func(t interface{}) string {
			return t.(*mockKrakenTrade).Cost
		},
		GetTime: func(t interface{}) interface{} {
			return t.(*mockKrakenTrade).Time
		},
		GetTradeID: func(t interface{}) interface{} {
			return t.(*mockKrakenTrade).TradeID
		},
		UpdateSize: func(t interface{}, v string) {
			t.(*mockKrakenTrade).Vol = v
		},
		UpdateFee: func(t interface{}, v string) {
			t.(*mockKrakenTrade).Fee = v
		},
		UpdateCost: func(t interface{}, v string) {
			t.(*mockKrakenTrade).Cost = v
		},
		UpdateTime: func(t interface{}, v interface{}) {
			if time, ok := v.(float64); ok {
				t.(*mockKrakenTrade).Time = time
			}
		},
		UpdateTradeID: func(t interface{}, v interface{}) {
			t.(*mockKrakenTrade).TradeID = v
		},
	}
}

func TestIsEarlierTime(t *testing.T) {
	tests := []struct {
		name         string
		currentTime  interface{}
		earliestTime interface{}
		expected     bool
	}{
		{
			name:         "Kraken: current earlier than earliest",
			currentTime:  float64(1640000000.0),
			earliestTime: float64(1640000100.0),
			expected:     true,
		},
		{
			name:         "Kraken: current later than earliest",
			currentTime:  float64(1640000100.0),
			earliestTime: float64(1640000000.0),
			expected:     false,
		},
		{
			name:         "Hyperliquid: current earlier than earliest",
			currentTime:  int64(1640000000000),
			earliestTime: int64(1640000100000),
			expected:     true,
		},
		{
			name:         "Hyperliquid: current later than earliest",
			currentTime:  int64(1640000100000),
			earliestTime: int64(1640000000000),
			expected:     false,
		},
		{
			name:         "Same time",
			currentTime:  int64(1640000000000),
			earliestTime: int64(1640000000000),
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEarlierTime(tt.currentTime, tt.earliestTime)
			if result != tt.expected {
				t.Errorf("isEarlierTime(%v, %v) = %v, want %v", tt.currentTime, tt.earliestTime, result, tt.expected)
			}
		})
	}
}
