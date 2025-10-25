package trades

import (
	"github.com/zyriu/portfolio/backend/helpers/grist"
)

func UpdateBookEntry(trade grist.Trade, entry grist.BookEntry) (grist.Trade, grist.BookEntry) {
	if trade.Market == "Spot" {
		return updateSpot(trade, entry)
	}

	return updateFutures(trade, entry)
}

func updateSpot(trade grist.Trade, entry grist.BookEntry) (grist.Trade, grist.BookEntry) {
	// This is for Spot positions only. No shorts, so position size cannot be negative.
	qty := trade.OrderSize
	price := trade.Price

	pos := entry.PositionSize
	avg := entry.AveragePrice
	costBasis := entry.CostBasis

	realizedPnL := 0.0
	var newPos, newAvg, newCostBasis float64

	switch trade.Direction {
	case "Buy":
		// Increase/incept position, average up/down
		newPos = pos + qty
		newCostBasis = costBasis + price*qty
		if newPos != 0 {
			newAvg = newCostBasis / newPos
		} else {
			newAvg = 0
			newCostBasis = 0
		}
		// No realized PnL for buys in spot, as position is only opened or increased
	case "Sell":
		// Decrease position, calculate realized PnL
		if qty > pos {
			// Selling more than you have -- not allowed in spot, clamp qty
			qty = pos
		}
		newPos = pos - qty
		realizedPnL = (price - avg) * qty
		newCostBasis = avg * newPos
		if newPos != 0 {
			newAvg = avg
		} else {
			newAvg = 0
			newCostBasis = 0
		}
	}

	// Ensure position size is never negative in spot
	if newPos < 0 {
		newPos = 0
		newAvg = 0
		newCostBasis = 0
	}

	entry.PositionSize = newPos
	entry.AveragePrice = newAvg
	entry.CostBasis = newCostBasis

	trade.PnL = realizedPnL
	trade.OrderValue = price * qty // OrderValue, unless it's a forbidden sell (qty clamped above)

	return trade, entry
}

func updateFutures(trade grist.Trade, entry grist.BookEntry) (grist.Trade, grist.BookEntry) {
	// This is for Futures positions. Position size can be negative (naked short).
	qty := trade.OrderSize
	price := trade.Price

	pos := entry.PositionSize
	avg := entry.AveragePrice
	costBasis := entry.CostBasis

	realizedPnL := 0.0
	var newPos, newAvg, newCostBasis float64

	switch trade.Direction {
	case "Buy":
		if pos < 0 {
			// Reducing or closing a short position, maybe flipping to long
			if qty <= -pos {
				// Just reducing (not flipping)
				realizedPnL = (avg - price) * qty
				newPos = pos + qty
				newCostBasis = avg * newPos
				newAvg = avg
				if newPos == 0 {
					newAvg = 0
					newCostBasis = 0
				}
			} else {
				// Flipping from short to long: 1) close short, 2) open long for remainder
				realizedPnL = (avg - price) * (-pos)
				longQty := qty + pos // qty > -pos, so this is positive
				newPos = longQty
				newCostBasis = price * newPos
				if newPos != 0 {
					newAvg = price
				} else {
					newAvg = 0
					newCostBasis = 0
				}
			}
		} else {
			// Adding to or opening a long position, just average up/down
			newPos = pos + qty
			newCostBasis = costBasis + price*qty
			if newPos != 0 {
				newAvg = newCostBasis / newPos
			} else {
				newAvg = 0
				newCostBasis = 0
			}
		}
	case "Sell":
		if pos > 0 {
			// Reducing or closing a long position, maybe flipping to short
			if qty <= pos {
				// Just reducing (not flipping)
				realizedPnL = (price - avg) * qty
				newPos = pos - qty
				newCostBasis = avg * newPos
				newAvg = avg
				if newPos == 0 {
					newAvg = 0
					newCostBasis = 0
				}
			} else {
				// Flipping from long to short: 1) close long, 2) open short for remainder
				realizedPnL = (price - avg) * pos
				shortQty := qty - pos // qty > pos, so this is positive
				newPos = -shortQty
				newCostBasis = price * newPos // newPos negative, costBasis negative
				if newPos != 0 {
					newAvg = price
				} else {
					newAvg = 0
					newCostBasis = 0
				}
			}
		} else {
			// Adding to or opening a short position, just average up/down
			newPos = pos - qty
			newCostBasis = costBasis - price*qty
			if newPos != 0 {
				newAvg = newCostBasis / newPos
			} else {
				newAvg = 0
				newCostBasis = 0
			}
		}
	}

	// No constraint: newPos can be negative (short), positive (long), or zero (flat)

	entry.PositionSize = newPos
	entry.AveragePrice = newAvg
	entry.CostBasis = newCostBasis

	trade.PnL = realizedPnL
	trade.OrderValue = price * qty

	return trade, entry
}
