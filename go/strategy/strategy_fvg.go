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
		panic("Could not load time zone: " + err.Error())
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

// Variablar som behövs för fvg
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
	//Kolla efter long
	if candle3.Low > candle1.High {
		return FVG{
			Top:       candle3.Low,
			Bottom:    candle1.High,
			IsBullish: true,
			Valid:     true,
		}
	}
	//Kolla efter short 
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

	//Konvertera int64 till time.Time
	barTime := time.Unix(bar.Timestamp, 0)
	nyTime := barTime.In(s.nyZone)

	if !isEarlySession(nyTime) {
		return types.Signal{Action: "HOLD", Reason: "outside early session"}
	}
	//Bestäm trenden
	currentBar := history[len(history)-1]
	sma := engine.SMA(history, s.SMAPeriod)
	isUptrend := currentBar.Close > sma

	for i := len(history) - 10; i < len(history)-3; i++ {
		if i < 0 {
			continue
		}
		//Kolla efter FVG senaste 10 bars
		fvg := detectFVG(history[i : i+3])
		if !fvg.Valid {
			continue
		}
		//Är priset i gapet
		priceInGap := currentBar.Low <= fvg.Top && currentBar.High >= fvg.Bottom

		if !priceInGap {
			continue
		}
		//Kolla efter bullish FVG
		if fvg.IsBullish && isUptrend {
			isRed := currentBar.Close < currentBar.Open
			isGreen := currentBar.Close > currentBar.Open

			closeInGap := currentBar.Close >= fvg.Bottom && currentBar.Close <= fvg.Top
			openInGap := currentBar.Open >= fvg.Bottom && currentBar.Open <= fvg.Top
			closeBelowGap := currentBar.Close < fvg.Bottom

			// Invalidation: röd bar som stänger under gapet → gapet är trasigt, gå vidare
			if isRed && closeBelowGap {
				continue
			}

			// Två giltiga setup-varianter
			validRedRetest := isRed && closeInGap    // röd som höll sig kvar i gapet
			validGreenBounce := isGreen && openInGap // grön som studsade ur gapet
			//Kolla att någon av de stämmer
			if validRedRetest || validGreenBounce {
				stopLoss := fvg.Bottom - 0.01
				risk := currentBar.Close - stopLoss
				target := currentBar.Close + (risk * s.RiskReward)
				return types.Signal{
					Action:   "BUY",
					Size:     1.0,
					StopLoss: stopLoss,
					Target:   target,
					Reason:   "bullish FVG hold/bounce in uptrend",
				}
			}
		}

		//Kolla efter bearish FVG
		if !fvg.IsBullish && !isUptrend {
			isRed := currentBar.Close < currentBar.Open
			isGreen := currentBar.Close > currentBar.Open

			closeInGap := currentBar.Close >= fvg.Bottom && currentBar.Close <= fvg.Top
			openInGap := currentBar.Open >= fvg.Bottom && currentBar.Open <= fvg.Top
			closeAboveGap := currentBar.Close > fvg.Top

			// Invalidation: grön bar som stänger över gapet → gapet är trasigt
			if isGreen && closeAboveGap {
				continue
			}

			validGreenRetest := isGreen && closeInGap // grön som höll sig kvar i gapet
			validRedRejection := isRed && openInGap   // röd som studsade ner ur gapet
			//Kolla att någon av de stämmer
			if validGreenRetest || validRedRejection {
				stopLoss := fvg.Top + 0.01
				risk := stopLoss - currentBar.Close
				target := currentBar.Close - (risk * s.RiskReward)
				return types.Signal{
					Action:   "SELL",
					Size:     1.0,
					StopLoss: stopLoss,
					Target:   target,
					Reason:   "bearish FVG hold/rejection in downtrend",
				}
			}
		}
	}
	return types.Signal{Action: "HOLD", Reason: "no valid FVG setup"}
}
