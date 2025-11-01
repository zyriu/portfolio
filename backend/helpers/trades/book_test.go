package trades

import (
	"math"
	"testing"

	"github.com/zyriu/portfolio/backend/helpers/grist"
)

// Helper function to create a spot trade
func spotTrade(direction string, price, size float64) grist.Trade {
	return grist.Trade{
		Direction: direction,
		Price:     price,
		OrderSize: size,
		Market:    "Spot",
	}
}

// Helper function to create a futures trade
func futuresTrade(direction string, price, size float64) grist.Trade {
	return grist.Trade{
		Direction: direction,
		Price:     price,
		OrderSize: size,
		Market:    "Futures",
	}
}

// Helper function to create a book entry
func entry(exchange, ticker, market string, avgPrice, posSize, costBasis float64) grist.BookEntry {
	return grist.BookEntry{
		Exchange:     exchange,
		Ticker:       ticker,
		Market:       market,
		AveragePrice: avgPrice,
		PositionSize: posSize,
		CostBasis:    costBasis,
	}
}

// Helper to check if two floats are approximately equal
func approxEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// Test UpdateBookEntry routing
func TestUpdateBookEntry_Routing(t *testing.T) {
	t.Run("routes to spot for Spot market", func(t *testing.T) {
		trade := spotTrade("Buy", 100, 10)
		entry := entry("exchange", "BTC", "Spot", 0, 0, 0)

		updatedTrade, updatedEntry := UpdateBookEntry(trade, entry)

		if updatedEntry.Market != "Spot" {
			t.Errorf("expected Spot market, got %s", updatedEntry.Market)
		}
		if updatedEntry.PositionSize != 10 {
			t.Errorf("expected position size 10, got %f", updatedEntry.PositionSize)
		}
		if updatedTrade.PnL != 0 {
			t.Errorf("expected PnL 0 for buy, got %f", updatedTrade.PnL)
		}
	})

	t.Run("routes to futures for Futures market", func(t *testing.T) {
		trade := futuresTrade("Buy", 100, 10)
		entry := entry("exchange", "BTC", "Futures", 0, 0, 0)

		_, updatedEntry := UpdateBookEntry(trade, entry)

		if updatedEntry.Market != "Futures" {
			t.Errorf("expected Futures market, got %s", updatedEntry.Market)
		}
		if updatedEntry.PositionSize != 10 {
			t.Errorf("expected position size 10, got %f", updatedEntry.PositionSize)
		}
	})
}

// SPOT MARKET TESTS

func TestUpdateSpot_BuyFromZero(t *testing.T) {
	trade := spotTrade("Buy", 100, 10)
	entry := entry("exchange", "BTC", "Spot", 0, 0, 0)

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	if updatedEntry.PositionSize != 10 {
		t.Errorf("expected position size 10, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 100 {
		t.Errorf("expected average price 100, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 1000 {
		t.Errorf("expected cost basis 1000, got %f", updatedEntry.CostBasis)
	}
	if updatedTrade.PnL != 0 {
		t.Errorf("expected PnL 0, got %f", updatedTrade.PnL)
	}
	if updatedTrade.OrderValue != 1000 {
		t.Errorf("expected order value 1000, got %f", updatedTrade.OrderValue)
	}
}

func TestUpdateSpot_BuyMultipleTimes(t *testing.T) {
	// First buy
	entry := entry("exchange", "BTC", "Spot", 0, 0, 0)
	trade1 := spotTrade("Buy", 100, 10)
	_, entry = updateSpot(trade1, entry)

	// Second buy at different price
	trade2 := spotTrade("Buy", 150, 5)
	_, entry = updateSpot(trade2, entry)

	expectedPos := 15.0
	expectedCostBasis := 1000.0 + 750.0 // 100*10 + 150*5
	expectedAvg := expectedCostBasis / expectedPos

	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, entry.AveragePrice)
	}
	if !approxEqual(entry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, entry.CostBasis)
	}
}

func TestUpdateSpot_SellPartial(t *testing.T) {
	// Start with position
	entry := entry("exchange", "BTC", "Spot", 100, 10, 1000)
	trade := spotTrade("Sell", 150, 5)

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	expectedPos := 5.0
	expectedPnL := (150.0 - 100.0) * 5.0 // 250
	expectedCostBasis := 100.0 * 5.0     // 500
	expectedAvg := 100.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateSpot_SellAll(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 100, 10, 1000)
	trade := spotTrade("Sell", 150, 10)

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	if !approxEqual(updatedEntry.PositionSize, 0, 0.01) {
		t.Errorf("expected position size 0, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 0 {
		t.Errorf("expected average price 0, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 0 {
		t.Errorf("expected cost basis 0, got %f", updatedEntry.CostBasis)
	}
	expectedPnL := (150.0 - 100.0) * 10.0 // 500
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateSpot_SellMoreThanPosition(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 100, 10, 1000)
	trade := spotTrade("Sell", 150, 15) // Trying to sell more than owned

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	// Should cap at position size
	if !approxEqual(updatedEntry.PositionSize, 0, 0.01) {
		t.Errorf("expected position size 0, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 0 {
		t.Errorf("expected average price 0, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 0 {
		t.Errorf("expected cost basis 0, got %f", updatedEntry.CostBasis)
	}
	// PnL should be calculated on actual sold amount (10, not 15)
	expectedPnL := (150.0 - 100.0) * 10.0 // 500
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
	// Order value should reflect actual trade size (10 * 150 = 1500)
	expectedOrderValue := 150.0 * 10.0 // capped to position size
	if !approxEqual(updatedTrade.OrderValue, expectedOrderValue, 0.01) {
		t.Errorf("expected order value %f, got %f", expectedOrderValue, updatedTrade.OrderValue)
	}
}

func TestUpdateSpot_SellFromZero(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 0, 0, 0)
	trade := spotTrade("Sell", 150, 10)

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	if updatedEntry.PositionSize != 0 {
		t.Errorf("expected position size 0, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 0 {
		t.Errorf("expected average price 0, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 0 {
		t.Errorf("expected cost basis 0, got %f", updatedEntry.CostBasis)
	}
	if updatedTrade.PnL != 0 {
		t.Errorf("expected PnL 0 (can't realize PnL from zero), got %f", updatedTrade.PnL)
	}
}

func TestUpdateSpot_BuySellSequence(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 0, 0, 0)

	// Buy 10 at 100
	trade1 := spotTrade("Buy", 100, 10)
	_, entry = updateSpot(trade1, entry)
	if entry.PositionSize != 10 || entry.AveragePrice != 100 {
		t.Errorf("after first buy: pos=%f avg=%f", entry.PositionSize, entry.AveragePrice)
	}

	// Sell 3 at 120
	trade2 := spotTrade("Sell", 120, 3)
	_, entry = updateSpot(trade2, entry)
	if !approxEqual(entry.PositionSize, 7, 0.01) {
		t.Errorf("after sell: expected pos 7, got %f", entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, 100, 0.01) {
		t.Errorf("after sell: expected avg 100, got %f", entry.AveragePrice)
	}

	// Buy 5 more at 110
	trade3 := spotTrade("Buy", 110, 5)
	_, entry = updateSpot(trade3, entry)
	expectedPos := 12.0
	expectedCostBasis := 100.0*7.0 + 110.0*5.0 // 1250
	expectedAvg := expectedCostBasis / expectedPos
	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("after second buy: expected pos %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("after second buy: expected avg %f, got %f", expectedAvg, entry.AveragePrice)
	}
}

func TestUpdateSpot_NegativePositionCorrection(t *testing.T) {
	// Test the safeguard that prevents negative positions
	entry := entry("exchange", "BTC", "Spot", 100, 5, 500)
	trade := spotTrade("Sell", 150, 10) // Selling more than owned

	_, updatedEntry := updateSpot(trade, entry)

	// Should be capped at 0, not negative
	if updatedEntry.PositionSize < 0 {
		t.Errorf("position should not be negative, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.PositionSize != 0 {
		t.Errorf("expected position size 0, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 0 {
		t.Errorf("expected average price 0, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 0 {
		t.Errorf("expected cost basis 0, got %f", updatedEntry.CostBasis)
	}
}

// FUTURES MARKET TESTS

func TestUpdateFutures_BuyFromZero(t *testing.T) {
	trade := futuresTrade("Buy", 100, 10)
	entry := entry("exchange", "BTC", "Futures", 0, 0, 0)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	if updatedEntry.PositionSize != 10 {
		t.Errorf("expected position size 10, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 100 {
		t.Errorf("expected average price 100, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 1000 {
		t.Errorf("expected cost basis 1000, got %f", updatedEntry.CostBasis)
	}
	if updatedTrade.PnL != 0 {
		t.Errorf("expected PnL 0, got %f", updatedTrade.PnL)
	}
}

func TestUpdateFutures_SellFromZero(t *testing.T) {
	trade := futuresTrade("Sell", 100, 10)
	entry := entry("exchange", "BTC", "Futures", 0, 0, 0)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	if updatedEntry.PositionSize != -10 {
		t.Errorf("expected position size -10, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 100 {
		t.Errorf("expected average price 100, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != -1000 {
		t.Errorf("expected cost basis -1000, got %f", updatedEntry.CostBasis)
	}
	if updatedTrade.PnL != 0 {
		t.Errorf("expected PnL 0, got %f", updatedTrade.PnL)
	}
}

func TestUpdateFutures_BuyMoreLong(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, 10, 1000)
	trade := futuresTrade("Buy", 150, 5)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	expectedPos := 15.0
	expectedCostBasis := 1000.0 + 750.0 // 1750
	expectedAvg := expectedCostBasis / expectedPos

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if updatedTrade.PnL != 0 {
		t.Errorf("expected PnL 0 (adding to long), got %f", updatedTrade.PnL)
	}
}

func TestUpdateFutures_SellPartialLong(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, 10, 1000)
	trade := futuresTrade("Sell", 150, 5)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	expectedPos := 5.0
	expectedPnL := (150.0 - 100.0) * 5.0 // 250
	expectedCostBasis := 100.0 * 5.0     // 500
	expectedAvg := 100.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateFutures_SellMoreLong_FlipsToShort(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, 10, 1000)
	trade := futuresTrade("Sell", 150, 15) // Selling more than long position

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	// Should close long and open short
	expectedPos := -5.0                   // 10 long - 15 sold = -5 short
	expectedPnL := (150.0 - 100.0) * 10.0 // Realized on closing the 10 long
	expectedCostBasis := 150.0 * (-5.0)   // New short position cost basis
	expectedAvg := 150.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateFutures_SellMoreLong_ClosesExactly(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, 10, 1000)
	trade := futuresTrade("Sell", 150, 10) // Selling exactly the position

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	expectedPos := 0.0
	expectedPnL := (150.0 - 100.0) * 10.0 // 500
	expectedCostBasis := 0.0
	expectedAvg := 0.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != expectedAvg {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != expectedCostBasis {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateFutures_SellMoreShort(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, -10, -1000)
	trade := futuresTrade("Sell", 120, 5)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	expectedPos := -15.0
	expectedCostBasis := -1000.0 - 600.0 // -1600
	expectedAvg := expectedCostBasis / expectedPos

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if updatedTrade.PnL != 0 {
		t.Errorf("expected PnL 0 (adding to short), got %f", updatedTrade.PnL)
	}
}

func TestUpdateFutures_BuyPartialShort(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, -10, -1000)
	trade := futuresTrade("Buy", 80, 5)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	expectedPos := -5.0
	expectedPnL := (100.0 - 80.0) * 5.0 // 100 (profitable to close short at lower price)
	expectedCostBasis := 100.0 * (-5.0) // -500
	expectedAvg := 100.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateFutures_BuyMoreShort_FlipsToLong(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, -10, -1000)
	trade := futuresTrade("Buy", 80, 15) // Buying more than short position

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	// Should close short and open long
	expectedPos := 5.0                   // -10 short + 15 bought = 5 long
	expectedPnL := (100.0 - 80.0) * 10.0 // Realized on closing the 10 short
	expectedCostBasis := 80.0 * 5.0      // New long position cost basis
	expectedAvg := 80.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if !approxEqual(updatedEntry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateFutures_BuyMoreShort_ClosesExactly(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, -10, -1000)
	trade := futuresTrade("Buy", 80, 10) // Buying exactly the position

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	expectedPos := 0.0
	expectedPnL := (100.0 - 80.0) * 10.0 // 200
	expectedCostBasis := 0.0
	expectedAvg := 0.0

	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != expectedAvg {
		t.Errorf("expected average price %f, got %f", expectedAvg, updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != expectedCostBasis {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, updatedEntry.CostBasis)
	}
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateFutures_ComplexSequence(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 0, 0, 0)

	// Buy 10 at 100 (long)
	trade1 := futuresTrade("Buy", 100, 10)
	_, entry = updateFutures(trade1, entry)
	if entry.PositionSize != 10 || entry.AveragePrice != 100 {
		t.Errorf("after first buy: pos=%f avg=%f", entry.PositionSize, entry.AveragePrice)
	}

	// Sell 5 at 110 (reduce long)
	trade2 := futuresTrade("Sell", 110, 5)
	_, entry = updateFutures(trade2, entry)
	if !approxEqual(entry.PositionSize, 5, 0.01) {
		t.Errorf("after first sell: expected pos 5, got %f", entry.PositionSize)
	}

	// Sell 10 at 120 (flip to short)
	trade3 := futuresTrade("Sell", 120, 10)
	_, entry = updateFutures(trade3, entry)
	expectedPos := -5.0
	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("after second sell: expected pos %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, 120, 0.01) {
		t.Errorf("after second sell: expected avg 120, got %f", entry.AveragePrice)
	}

	// Buy 8 at 105 (flip back to long)
	trade4 := futuresTrade("Buy", 105, 8)
	_, entry = updateFutures(trade4, entry)
	expectedPos = 3.0 // -5 + 8
	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("after second buy: expected pos %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, 105, 0.01) {
		t.Errorf("after second buy: expected avg 105, got %f", entry.AveragePrice)
	}
}

// EDGE CASES

func TestUpdateSpot_ZeroQuantity(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 100, 10, 1000)
	trade := spotTrade("Buy", 100, 0)

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	if updatedEntry.PositionSize != 10 {
		t.Errorf("position should remain unchanged, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 100 {
		t.Errorf("average price should remain unchanged, got %f", updatedEntry.AveragePrice)
	}
	if updatedTrade.OrderValue != 0 {
		t.Errorf("order value should be 0, got %f", updatedTrade.OrderValue)
	}
}

func TestUpdateFutures_ZeroQuantity(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, 10, 1000)
	trade := futuresTrade("Buy", 100, 0)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	if updatedEntry.PositionSize != 10 {
		t.Errorf("position should remain unchanged, got %f", updatedEntry.PositionSize)
	}
	if updatedTrade.OrderValue != 0 {
		t.Errorf("order value should be 0, got %f", updatedTrade.OrderValue)
	}
}

func TestUpdateSpot_VerySmallQuantities(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 100, 10, 1000)
	trade := spotTrade("Buy", 100, 0.0001)

	_, updatedEntry := updateSpot(trade, entry)

	expectedPos := 10.0001
	if !approxEqual(updatedEntry.PositionSize, expectedPos, 0.00001) {
		t.Errorf("expected position size %f, got %f", expectedPos, updatedEntry.PositionSize)
	}
}

func TestUpdateFutures_ZeroPositionWithNonZeroAverage(t *testing.T) {
	// Edge case: position is 0 but average/cost basis might be stale
	entry := entry("exchange", "BTC", "Futures", 100, 0, 0)
	trade := futuresTrade("Buy", 150, 5)

	_, updatedEntry := updateFutures(trade, entry)

	if !approxEqual(updatedEntry.PositionSize, 5, 0.01) {
		t.Errorf("expected position size 5, got %f", updatedEntry.PositionSize)
	}
	if !approxEqual(updatedEntry.AveragePrice, 150, 0.01) {
		t.Errorf("expected average price 150, got %f", updatedEntry.AveragePrice)
	}
}

func TestUpdateFutures_LongToShortExactFlip(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 100, 5, 500)
	trade := futuresTrade("Sell", 120, 5)

	updatedTrade, updatedEntry := updateFutures(trade, entry)

	if updatedEntry.PositionSize != 0 {
		t.Errorf("expected position size 0, got %f", updatedEntry.PositionSize)
	}
	if updatedEntry.AveragePrice != 0 {
		t.Errorf("expected average price 0, got %f", updatedEntry.AveragePrice)
	}
	if updatedEntry.CostBasis != 0 {
		t.Errorf("expected cost basis 0, got %f", updatedEntry.CostBasis)
	}
	expectedPnL := (120.0 - 100.0) * 5.0 // 100
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
}

func TestUpdateSpot_MultipleRapidTrades(t *testing.T) {
	// Simulate rapid trades that might cause duplicate entries
	entry := entry("exchange", "BTC", "Spot", 0, 0, 0)

	// Rapid sequence of buys
	for i := 0; i < 10; i++ {
		trade := spotTrade("Buy", float64(100+i), 1)
		_, entry = updateSpot(trade, entry)
	}

	expectedPos := 10.0
	expectedCostBasis := 0.0
	for i := 0; i < 10; i++ {
		expectedCostBasis += float64(100 + i)
	}
	expectedAvg := expectedCostBasis / expectedPos

	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("expected position size %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("expected average price %f, got %f", expectedAvg, entry.AveragePrice)
	}
	if !approxEqual(entry.CostBasis, expectedCostBasis, 0.01) {
		t.Errorf("expected cost basis %f, got %f", expectedCostBasis, entry.CostBasis)
	}

	// Now rapid sequence of sells
	for i := 0; i < 5; i++ {
		trade := spotTrade("Sell", float64(120+i), 2)
		_, entry = updateSpot(trade, entry)
	}

	if !approxEqual(entry.PositionSize, 0, 0.01) {
		t.Errorf("after all sells, expected position size 0, got %f", entry.PositionSize)
	}
	if entry.AveragePrice != 0 {
		t.Errorf("after all sells, expected average price 0, got %f", entry.AveragePrice)
	}
	if entry.CostBasis != 0 {
		t.Errorf("after all sells, expected cost basis 0, got %f", entry.CostBasis)
	}
}

func TestUpdateFutures_MultipleFlips(t *testing.T) {
	// Test multiple position flips in a row
	entry := entry("exchange", "BTC", "Futures", 0, 0, 0)

	// Long to short
	trade1 := futuresTrade("Buy", 100, 10)
	_, entry = updateFutures(trade1, entry)

	trade2 := futuresTrade("Sell", 110, 15)
	_, entry = updateFutures(trade2, entry)
	if !approxEqual(entry.PositionSize, -5, 0.01) {
		t.Errorf("after flip to short: expected pos -5, got %f", entry.PositionSize)
	}

	// Short to long
	trade3 := futuresTrade("Buy", 105, 10)
	_, entry = updateFutures(trade3, entry)
	if !approxEqual(entry.PositionSize, 5, 0.01) {
		t.Errorf("after flip to long: expected pos 5, got %f", entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, 105, 0.01) {
		t.Errorf("after flip to long: expected avg 105, got %f", entry.AveragePrice)
	}
}

func TestUpdateSpot_NegativePnL(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 100, 10, 1000)
	trade := spotTrade("Sell", 80, 5) // Selling at loss

	updatedTrade, updatedEntry := updateSpot(trade, entry)

	expectedPnL := (80.0 - 100.0) * 5.0 // -100 (loss)
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}
	if !approxEqual(updatedEntry.PositionSize, 5, 0.01) {
		t.Errorf("expected position size 5, got %f", updatedEntry.PositionSize)
	}
}

func TestUpdateFutures_NegativePnL(t *testing.T) {
	entry1 := entry("exchange", "BTC", "Futures", 100, 10, 1000)
	trade := futuresTrade("Sell", 80, 5) // Selling long at loss

	updatedTrade, _ := updateFutures(trade, entry1)

	expectedPnL := (80.0 - 100.0) * 5.0 // -100 (loss)
	if !approxEqual(updatedTrade.PnL, expectedPnL, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL, updatedTrade.PnL)
	}

	// Test short closing at loss
	entry2 := entry("exchange", "BTC", "Futures", 100, -10, -1000)
	trade2 := futuresTrade("Buy", 120, 5) // Buying short back at higher price (loss)

	updatedTrade2, _ := updateFutures(trade2, entry2)
	expectedPnL2 := (100.0 - 120.0) * 5.0 // -100 (loss)
	if !approxEqual(updatedTrade2.PnL, expectedPnL2, 0.01) {
		t.Errorf("expected PnL %f, got %f", expectedPnL2, updatedTrade2.PnL)
	}
}

// Test for potential duplicate entry scenarios
func TestUpdateSpot_RepeatedOperations(t *testing.T) {
	entry := entry("exchange", "BTC", "Spot", 0, 0, 0)

	// Same trade applied twice (simulating duplicate processing)
	trade := spotTrade("Buy", 100, 10)

	_, entry = updateSpot(trade, entry)

	// Apply same trade again
	_, entry = updateSpot(trade, entry)

	// Should accumulate, not create duplicate
	expectedPos := 20.0
	expectedAvg := 100.0 // Same price
	expectedCost := 2000.0

	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("duplicate trade: expected pos %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("duplicate trade: expected avg %f, got %f", expectedAvg, entry.AveragePrice)
	}
	if !approxEqual(entry.CostBasis, expectedCost, 0.01) {
		t.Errorf("duplicate trade: expected cost %f, got %f", expectedCost, entry.CostBasis)
	}
}

func TestUpdateFutures_RepeatedOperations(t *testing.T) {
	entry := entry("exchange", "BTC", "Futures", 0, 0, 0)

	// Same trade applied multiple times
	trade := futuresTrade("Sell", 100, 5)

	for i := 0; i < 3; i++ {
		_, entry = updateFutures(trade, entry)
	}

	expectedPos := -15.0
	expectedAvg := 100.0
	expectedCost := -1500.0

	if !approxEqual(entry.PositionSize, expectedPos, 0.01) {
		t.Errorf("repeated trades: expected pos %f, got %f", expectedPos, entry.PositionSize)
	}
	if !approxEqual(entry.AveragePrice, expectedAvg, 0.01) {
		t.Errorf("repeated trades: expected avg %f, got %f", expectedAvg, entry.AveragePrice)
	}
	if !approxEqual(entry.CostBasis, expectedCost, 0.01) {
		t.Errorf("repeated trades: expected cost %f, got %f", expectedCost, entry.CostBasis)
	}
}
