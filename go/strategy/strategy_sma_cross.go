package strategies

import (
	"time"
	"trading-bot/engine"
	"trading-bot/types"
)

// So that we can create different SMA strategies
type SMACrossStrategy struct {
	SMAPeriod       int
	VolumePeriod    int
	VolumeThreshold float64
	SwingLookback   int
	RiskReward      float64
	nyZone          *time.Location
}

// First SMA strategy
func NewSMACrossStrategy() *SMACrossStrategy {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic("Could not load time zone: " + err.Error())
	}
	return &SMACrossStrategy{
		SMAPeriod:       100,
		VolumePeriod:    20,
		VolumeThreshold: 1.2,
		SwingLookback:   10,
		RiskReward:      2.0,
		nyZone:          loc,
	}
}

// Returns the name
func (s *SMACrossStrategy) Name() string {
	return "SMA_Cross_Volume"
}

// Checks that we have enough data
func (s *SMACrossStrategy) OnBar(bar types.Bar, history []types.Bar) types.Signal {
	if len(history) < s.SMAPeriod+1 {
		return types.Signal{Action: "HOLD", Reason: "insufficient data"}
	} //Checks that the time is right

	//converts from int64 to time.Time
	barTime := time.Unix(bar.Timestamp, 0)
	nyTime := barTime.In(s.nyZone)

	if !isEarlySession(nyTime) {
		return types.Signal{Action: "HOLD", Reason: "outside early session"}
	}

	//Check SMA on different bars to know when it crosses
	currentSMA := engine.SMA(history, s.SMAPeriod)
	prevBars := history[:len(history)-1]
	prevSMA := engine.SMA(prevBars, s.SMAPeriod)

	currentBar := history[len(history)-1]
	prevBar := history[len(history)-2]
	//Calculates average volume
	avgVol := engine.AvgVolume(history[:len(history)-1], s.VolumePeriod)
	if float64(currentBar.Volume) < avgVol*s.VolumeThreshold {
		return types.Signal{Action: "HOLD", Reason: "volume too low"}
	}

	//Enter long trade
	if prevBar.Close < prevSMA && currentBar.Close > currentSMA { //We have crossed up through the SMA
		swingLow := engine.SwingLow(history, s.SwingLookback)
		stopLoss := swingLow - 0.01 //StopLoss under swing low
		risk := currentBar.Close - stopLoss
		target := currentBar.Close + (risk * s.RiskReward) //Target 2x the risk

		return types.Signal{
			Action:   "BUY",
			Size:     1.0,
			StopLoss: stopLoss,
			Target:   target,
			Reason:   "SMA cross up with volume confirmation",
		}
	}

	//Enter short trade, same as long
	if prevBar.Close > prevSMA && currentBar.Close < currentSMA {
		swingHigh := engine.SwingHigh(history, s.SwingLookback)
		stopLoss := swingHigh + 0.01
		risk := stopLoss - currentBar.Close
		target := currentBar.Close - (risk * s.RiskReward)

		return types.Signal{
			Action:   "SELL",
			Size:     1.0,
			StopLoss: stopLoss,
			Target:   target,
			Reason:   "SMA cross down with volume confirmation",
		}
	}
	return types.Signal{Action: "HOLD", Reason: "no cross detected"}
}
