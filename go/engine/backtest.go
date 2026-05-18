package engine

import (
	"trading-bot/types"
)

// Strategy är ett internt gränssnitt för att undvika cirkulära beroenden
type Strategy interface {
	Name() string
	OnBar(bar types.Bar, history []types.Bar) types.Signal
}

func RunBacktest(ticker string, data []types.Bar, strat Strategy) []types.Trade {
	var trades []types.Trade
	var inPosition bool
	var entryPrice float64
	var entryTime int64
	var positionAction string
	var currentStopLoss float64
	var currentTarget float64

	var history []types.Bar
	maxHistory := 1000

	for _, bar := range data {
		history = append(history, bar)
		if len(history) > maxHistory {
			history = history[1:]
		}

		// Check exits if we are in a position
		if inPosition {
			var exitPrice float64
			var exited bool

			if positionAction == "BUY" {
				if bar.Low <= currentStopLoss {
					exitPrice = currentStopLoss
					exited = true
				} else if bar.High >= currentTarget {
					exitPrice = currentTarget
					exited = true
				}
			} else if positionAction == "SELL" {
				if bar.High >= currentStopLoss {
					exitPrice = currentStopLoss
					exited = true
				} else if bar.Low <= currentTarget {
					exitPrice = currentTarget
					exited = true
				}
			}

			if exited {
				var pnl float64
				if positionAction == "BUY" {
					pnl = exitPrice - entryPrice
				} else {
					pnl = entryPrice - exitPrice
				}

				trades = append(trades, types.Trade{
					Timestamp:  entryTime,
					CloseTime:  bar.Timestamp,
					Symbol:     ticker,
					Strategy:   strat.Name(),
					Action:     positionAction,
					Price:      entryPrice,
					EntryPrice: entryPrice,
					ClosePrice: exitPrice,
					ProfitLoss: pnl,
				})
				inPosition = false

				continue
			}

			// Simple logic: if we are in a position, we don't take new trades
			if inPosition {
				continue
			}
		}

		// Ask strategy for signal
		signal := strat.OnBar(bar, history)

		if signal.Action == "BUY" || signal.Action == "SELL" {
			inPosition = true
			positionAction = signal.Action
			entryPrice = bar.Close
			entryTime = bar.Timestamp
			currentStopLoss = signal.StopLoss
			currentTarget = signal.Target
		}
	}
	return trades
}
