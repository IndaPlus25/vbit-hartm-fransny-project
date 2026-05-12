package engine

import (
	"trading-bot/types"
)

func RunSimulation(data []types.Bar, strat types.Strategy) []types.Trade {
	var results []types.Trade
	var activeTrade *types.Trade
	var history []types.Bar

	strategyName := strat.Name()

	for _, bar := range data {
		if activeTrade != nil {
			if bar.Low <= activeTrade.StopLoss || bar.High >= activeTrade.Target {
				activeTrade.ClosePrice = bar.Close
				if activeTrade.Action == "BUY" {
					activeTrade.ProfitLoss = activeTrade.ClosePrice - activeTrade.EntryPrice
				} else if activeTrade.Action == "SELL" {
					activeTrade.ProfitLoss = activeTrade.EntryPrice - activeTrade.ClosePrice
				}
				results = append(results, *activeTrade)
				activeTrade = nil
			}
		}

		history = append(history, bar)
		if len(history) > 200 //måste inte vara 200 {
			history = history[1:]
		}
		if activeTrade == nil {
			signal := strat.OnBar(bar, history)

			if signal.Action == "BUY" || signal.Action == "SELL" {
				activeTrade = &types.Trade{
					StrategyName: strategyName,
					Action:       signal.Action,
					EntryPrice:   bar.Close,
					StopLoss:     signal.StopLoss,
					Target:       signal.Target,
				}
			}
		}
	}
	return results
}
