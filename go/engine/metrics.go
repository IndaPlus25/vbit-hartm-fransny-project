package engine

import (
	"trading-bot/types"
)

// Function to track mean price over a period
func SMA(bars []types.Bar, period int) float64 {
	if len(bars) < period { //Can't calculate SMA without sufficient amount of bars
		return 0
	}
	sum := 0.0
	for i := len(bars) - period; i < len(bars); i++ { //Goes through all bars
		sum += bars[i].Close //Add the closing price
	}
	return sum / float64(period) //Average price durring the period
}

// Average volume durring a period
func AvgVolume(bars []types.Bar, period int) float64 {
	if len(bars) < period {
		return 0
	}
	sum := int64(0)
	for i := len(bars) - period; i < len(bars); i++ {
		sum += bars[i].Volume
	}
	return float64(sum) / float64(period)
}

// Lowest price durring a period
func SwingLow(bars []types.Bar, lookback int) float64 {
	if len(bars) < lookback {
		lookback = len(bars)
	}
	lowest := bars[len(bars)-1].Low
	for i := len(bars) - lookback; i < len(bars); i++ {
		if bars[i].Low < lowest {
			lowest = bars[i].Low //Update the lowest value
		}
	}
	return lowest
}

// Highest price durring a period
func SwingHigh(bars []types.Bar, lookback int) float64 {
	if len(bars) < lookback {
		lookback = len(bars)
	}
	highest := bars[len(bars)-1].High
	for i := len(bars) - lookback; i < len(bars); i++ {
		if bars[i].High > highest {
			highest = bars[i].High
		}
	}
	return highest
}

// AggregateBars slår ihop flera små bars till färre stora bars
func AggregateBars(bars []types.Bar, factor int) []types.Bar {
	if factor <= 1 {
		return bars
	}
	var aggregated []types.Bar

	for i := 0; i < len(bars); i += factor {
		end := i + factor
		if end > len(bars) {
			break // ignorera om vi inte har en komplett grupp på slutet
		}

		chunk := bars[i:end]
		high := chunk[0].High
		low := chunk[0].Low
		vol := int64(0)

		for _, b := range chunk {
			if b.High > high {
				high = b.High
			}
			if b.Low < low {
				low = b.Low
			}
			vol += b.Volume
		}

		aggBar := types.Bar{
			Timestamp: chunk[0].Timestamp,
			Open:      chunk[0].Open,
			High:      high,
			Low:       low,
			Close:     chunk[len(chunk)-1].Close,
			Volume:    vol,
		}
		aggregated = append(aggregated, aggBar)
	}
	return aggregated
}
