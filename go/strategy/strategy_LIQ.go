package strategies

import (
	"time"
	"trading-bot/engine"
	"trading-bot/types"
)

// Liquidity sweep med multi-timeframe-logik.
// Swing-nivåer identifieras på den högre timeframe (HTF), entries triggas
// på den lägre timeframe (LTF) som motorn matar in via OnBar.
type LiquiditySweepStrategy struct {
	HTFFactor     int // hur många LTF-staplar = 1 HTF-stapel (t.ex. 3 för 5min→15min)
	SwingLookback int // antal HTF-staplar att leta swings i
	SwingMinAge   int // hoppa över de senaste N HTF-staplarna
	RiskReward    float64
	nyZone        *time.Location
}

// First sweep strategy
func NewLiquiditySweepStrategy() *LiquiditySweepStrategy {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	return &LiquiditySweepStrategy{
		HTFFactor:     3,  //5-min → 15-min
		SwingLookback: 20, //20 HTF-staplar = 5 timmar av 15-min data
		SwingMinAge:   2,  //2 HTF-staplar buffert
		RiskReward:    2.0,
		nyZone:        loc,
	}
}

// Returns the name
func (s *LiquiditySweepStrategy) Name() string {
	return "Liquidity_Sweep_MTF"
}

func (s *LiquiditySweepStrategy) OnBar(bar types.Bar, history []types.Bar) types.Signal {
	//Vi behöver tillräckligt LTF-historik för att kunna bygga HTF-staplar.
	//För att få SwingLookback + SwingMinAge HTF-staplar behöver vi
	//(SwingLookback + SwingMinAge) * HTFFactor LTF-staplar, plus de senaste 2 LTF-staplarna.
	minLTFBars := (s.SwingLookback+s.SwingMinAge)*s.HTFFactor + 2
	if len(history) < minLTFBars {
		return types.Signal{Action: "HOLD", Reason: "insufficient data"}
	}

	//converts int64 to time.Time
	barTime := time.Unix(bar.Timestamp, 0)
	nyTime := barTime.In(s.nyZone)

	if !isEarlySession(nyTime) {
		return types.Signal{Action: "HOLD", Reason: "outside early session"}
	}

	currentBar := history[len(history)-1] //LTF — entry/bekräftelse-stapel
	prevBar := history[len(history)-2]    //LTF — den potentiella sweep-stapeln

	//Bygg HTF-serien från ALLA LTF-staplar UTOM de två senaste.
	//Vi exkluderar dem så att HTF-nivåerna inte påverkas av den pågående sweepen
	//eller bekräftelsen — de ska vara etablerade i historiken.
	htfSource := history[:len(history)-2]
	htfBars := engine.AggregateBars(htfSource, s.HTFFactor)

	if len(htfBars) < s.SwingLookback+s.SwingMinAge {
		return types.Signal{Action: "HOLD", Reason: "insufficient HTF data"}
	}

	//Skippa de senaste SwingMinAge HTF-staplarna när vi söker nivåer
	htfPrior := htfBars[:len(htfBars)-s.SwingMinAge]
	swingHigh := engine.SwingHigh(htfPrior, s.SwingLookback)
	swingLow := engine.SwingLow(htfPrior, s.SwingLookback)

	//Sweep above HTF swing high: prev LTF-stapel wickade över men stängde under
	prevSweptHigh := prevBar.High > swingHigh && prevBar.Close < swingHigh
	currentBearish := currentBar.Close < currentBar.Open

	if prevSweptHigh && currentBearish {
		stopLoss := prevBar.High + 0.01
		risk := stopLoss - currentBar.Close
		target := currentBar.Close - (risk * s.RiskReward)

		return types.Signal{
			Action:   "SELL",
			Size:     1.0,
			StopLoss: stopLoss,
			Target:   target,
			Reason:   "liquidity sweep above HTF swing high, bearish confirmation",
		}
	}

	//Sweep below HTF swing low: prev LTF-stapel wickade under men stängde över
	prevSweptLow := prevBar.Low < swingLow && prevBar.Close > swingLow
	currentBullish := currentBar.Close > currentBar.Open

	if prevSweptLow && currentBullish {
		stopLoss := prevBar.Low - 0.01
		risk := currentBar.Close - stopLoss
		target := currentBar.Close + (risk * s.RiskReward)

		return types.Signal{
			Action:   "BUY",
			Size:     1.0,
			StopLoss: stopLoss,
			Target:   target,
			Reason:   "liquidity sweep below HTF swing low, bullish confirmation",
		}
	}

	return types.Signal{Action: "HOLD", Reason: "no sweep detected"}
}
