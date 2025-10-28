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
	qty := trade.OrderSize
	price := trade.Price

	pos := entry.PositionSize
	avg := entry.AveragePrice
	costBasis := entry.CostBasis

	realizedPnL := 0.0
	var newPos, newAvg, newCostBasis float64

	switch trade.Direction {
	case "Buy":
		newPos = pos + qty
		newCostBasis = costBasis + price*qty
		if newPos != 0 {
			newAvg = newCostBasis / newPos
		} else {
			newAvg = 0
			newCostBasis = 0
		}
	case "Sell":
		qty = min(qty, pos)
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

	if newPos < 0 {
		newPos = 0
		newAvg = 0
		newCostBasis = 0
	}

	entry.PositionSize = newPos
	entry.AveragePrice = newAvg
	entry.CostBasis = newCostBasis

	trade.PnL = realizedPnL
	trade.OrderValue = price * qty

	return trade, entry
}

func updateFutures(trade grist.Trade, entry grist.BookEntry) (grist.Trade, grist.BookEntry) {
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
			if qty <= -pos {
				realizedPnL = (avg - price) * qty
				newPos = pos + qty
				newCostBasis = avg * newPos
				newAvg = avg
				if newPos == 0 {
					newAvg = 0
					newCostBasis = 0
				}
			} else {
				realizedPnL = (avg - price) * (-pos)
				longQty := qty + pos
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
			if qty <= pos {
				realizedPnL = (price - avg) * qty
				newPos = pos - qty
				newCostBasis = avg * newPos
				newAvg = avg
				if newPos == 0 {
					newAvg = 0
					newCostBasis = 0
				}
			} else {
				realizedPnL = (price - avg) * pos
				shortQty := qty - pos
				newPos = -shortQty
				newCostBasis = price * newPos
				if newPos != 0 {
					newAvg = price
				} else {
					newAvg = 0
					newCostBasis = 0
				}
			}
		} else {
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

	entry.PositionSize = newPos
	entry.AveragePrice = newAvg
	entry.CostBasis = newCostBasis

	trade.PnL = realizedPnL
	trade.OrderValue = price * qty

	return trade, entry
}
