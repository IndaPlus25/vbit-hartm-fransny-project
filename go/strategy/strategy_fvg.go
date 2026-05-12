package strategies

import (
	"time"
	"trading-bot/engine"
	"trading-bot/types"
)

type FVGStrategy struct {
	SMAPeriod  int
	RiskReward float64
	nyZone     *time.Location
}

func NewFVGStrategy() *FVGStrategy {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	return &FVGStrategy{
		SMAPeriod:  100,
		RiskReward: 1.5,
		nyZone:     loc,
	}
}

func (s *FVGStrategy) Name() string {
	return "FVG_Trend"
}

// Variables needed for fvg
type FVG struct {
	Top       float64
	Bottom    float64
	IsBullish bool
	Valid     bool
}

func detectFVG(bars []types.Bar) FVG {
	if len(bars) < 3 { //Need 3 bars
		return FVG{Valid: false}
	}

	candle1 := bars[len(bars)-3]
	candle3 := bars[len(bars)-1]
	//Look for long
	if candle3.Low > candle1.High {
		return FVG{
			Top:       candle3.Low,
			Bottom:    candle1.High,
			IsBullish: true,
			Valid:     true,
		}
	}
	//Look for short
	if candle3.High < candle1.Low {
		return FVG{
			Top:       candle1.Low,
			Bottom:    candle3.High,
			IsBullish: false,
			Valid:     true,
		}
	}

	return FVG{Valid: false}
}

func (s *FVGStrategy) OnBar(bar types.Bar, history []types.Bar) types.Signal {
	if len(history) < s.SMAPeriod+3 {
		return types.Signal{Action: "HOLD", Reason: "insufficient data"}
	}

	//converts int64 to time.Time
	barTime := time.Unix(bar.Timestamp, 0)
	nyTime := barTime.In(s.nyZone)

	if !isEarlySession(nyTime) {
		return types.Signal{Action: "HOLD", Reason: "outside early session"}
	}
	//Determine the trend
	currentBar := history[len(history)-1]
	sma := engine.SMA(history, s.SMAPeriod)
	isUptrend := currentBar.Close > sma

	for i := len(history) - 10; i < len(history)-3; i++ {
		if i < 0 {
			continue
		}
		//Look for fvg in the last 10 bars
		fvg := detectFVG(history[i : i+3])
		if !fvg.Valid {
			continue
		}
		//Is the price in the gap
		priceInGap := currentBar.Low <= fvg.Top && currentBar.High >= fvg.Bottom

		if !priceInGap {
			continue
		}
		//Look for bullish
		if fvg.IsBullish && isUptrend && currentBar.Close > currentBar.Open {
			stopLoss := fvg.Bottom - 0.01
			risk := currentBar.Close - stopLoss
			target := currentBar.Close + (risk * s.RiskReward)

			return types.Signal{
				Action:   "BUY",
				Size:     1.0,
				StopLoss: stopLoss,
				Target:   target,
				Reason:   "bullish FVG retest in uptrend",
			}
		}
		//Look for bearish
		if !fvg.IsBullish && !isUptrend && currentBar.Close < currentBar.Open {
			stopLoss := fvg.Top + 0.01
			risk := stopLoss - currentBar.Close
			target := currentBar.Close - (risk * s.RiskReward)

			return types.Signal{
				Action:   "SELL",
				Size:     1.0,
				StopLoss: stopLoss,
				Target:   target,
				Reason:   "bearish FVG retest in downtrend",
			}
		}
	}
	return types.Signal{Action: "HOLD", Reason: "no valid FVG setup"}
}
