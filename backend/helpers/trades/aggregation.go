package trades

import (
	"fmt"
	"strconv"
)

type AggregationKey struct {
	Asset     string
	TimeMin   int64
	Direction string
	Price     string
}

type AggregatedTrade struct {
	Trade           any
	Count           int
	AggregatedSize  float64
	AggregatedFee   float64
	AggregatedCost  float64
	EarliestTime    any
	EarliestTradeID any
}

type TradeConfig struct {
	GetAsset      func(any) string
	GetTimeMin    func(any) int64
	GetDirection  func(any) string
	GetPrice      func(any) string
	GetSize       func(any) string
	GetFee        func(any) string
	GetCost       func(any) string
	GetTime       func(any) any
	GetTradeID    func(any) any
	UpdateSize    func(any, string)
	UpdateFee     func(any, string)
	UpdateCost    func(any, string)
	UpdateTime    func(any, any)
	UpdateTradeID func(any, any)
}

func AggregateTrades(trades []any, config TradeConfig) map[AggregationKey]*AggregatedTrade {
	aggMap := make(map[AggregationKey]*AggregatedTrade)

	parseF := func(s string) float64 {
		val, _ := strconv.ParseFloat(s, 64)
		return val
	}
	formatF := func(f float64) string {
		return fmt.Sprintf("%.12g", f)
	}

	for _, trade := range trades {
		key := AggregationKey{
			Asset:     config.GetAsset(trade),
			TimeMin:   config.GetTimeMin(trade),
			Direction: config.GetDirection(trade),
			Price:     config.GetPrice(trade),
		}

		entry, exists := aggMap[key]
		if !exists {
			size := parseF(config.GetSize(trade))
			fee := parseF(config.GetFee(trade))
			cost := 0.0
			if config.GetCost != nil {
				costStr := config.GetCost(trade)
				if costStr != "" {
					cost = parseF(costStr)
				}
			}
			earliestTime := config.GetTime(trade)
			earliestTradeID := config.GetTradeID(trade)

			aggMap[key] = &AggregatedTrade{
				Trade:           trade,
				Count:           1,
				AggregatedSize:  size,
				AggregatedFee:   fee,
				AggregatedCost:  cost,
				EarliestTime:    earliestTime,
				EarliestTradeID: earliestTradeID,
			}
			continue
		}

		newSize := parseF(config.GetSize(trade))
		newFee := parseF(config.GetFee(trade))
		newCost := 0.0
		if config.GetCost != nil {
			costStr := config.GetCost(trade)
			if costStr != "" {
				newCost = parseF(costStr)
			}
		}

		entry.AggregatedSize += newSize
		entry.AggregatedFee += newFee
		entry.AggregatedCost += newCost
		entry.Count++

		currentTime := config.GetTime(trade)
		currentTradeID := config.GetTradeID(trade)

		if isEarlierTime(currentTime, entry.EarliestTime) {
			entry.Trade = trade
			entry.EarliestTime = currentTime
			entry.EarliestTradeID = currentTradeID
		}
	}

	for _, entry := range aggMap {
		config.UpdateSize(entry.Trade, formatF(entry.AggregatedSize))
		config.UpdateFee(entry.Trade, formatF(entry.AggregatedFee))
		if config.UpdateCost != nil {
			config.UpdateCost(entry.Trade, formatF(entry.AggregatedCost))
		}
		config.UpdateTime(entry.Trade, entry.EarliestTime)
		config.UpdateTradeID(entry.Trade, entry.EarliestTradeID)
	}

	return aggMap
}

func isEarlierTime(currentTime, earliestTime any) bool {
	current, ok1 := toComparableTime(currentTime)
	earliest, ok2 := toComparableTime(earliestTime)
	if !ok1 || !ok2 {
		return false
	}
	return current < earliest
}

func toComparableTime(t any) (int64, bool) {
	switch v := t.(type) {
	case float64:
		return int64(v * 1000), true
	case int64:
		return v, true
	default:
		return 0, false
	}
}
