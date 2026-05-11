package engine

import (
	"trading-bot/types"
)

func simulateTicker(ticker string, data []types.Bar) []types.Trade {
	var tradesForTicker []types.Trade
	var inPosition bool
	var buyPrice float64

	//kollar så vi har nog med SMA data (100)
	for i, tick := range data {
		if i < 100 {
			continue
		}
	}

	sma := calculateSMA(data[i-100 : i])
	//read data
	//loop over time
	//MATH functions

	return tradesForTicker
}
