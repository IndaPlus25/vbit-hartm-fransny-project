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
	
	var capital float64 = 10000.0 // Starting capital per strategy
	var shares float64

	var history []types.Bar
	maxHistory := 1000

	for _, bar := range data {
		history = append(history, bar)
		if len(history) > maxHistory {
			history = history[1:]
		}

		// Kollar exits om vi är i position
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
					pnl = (exitPrice - entryPrice) * shares
				} else {
					pnl = (entryPrice - exitPrice) * shares
				}
				
				capital += pnl // Update total capital with realized PnL

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

			// Simple logic: Är vi i en position, gör inga nya trades
			if inPosition {
				continue
			}
		}

		// Frågar strategi efter signal
		signal := strat.OnBar(bar, history)

		if signal.Action == "BUY" || signal.Action == "SELL" {
			inPosition = true
			positionAction = signal.Action
			entryPrice = bar.Close
			entryTime = bar.Timestamp
			currentStopLoss = signal.StopLoss
			currentTarget = signal.Target
			
			size := signal.Size
			if size <= 0 {
				size = 1.0 // Default to using 100% of capital if not specified
			}
			
			investedCapital := capital * size
			shares = investedCapital / entryPrice
		}
	}
	return trades
}
