package strategies

import (
	"time"
	"trading-bot/engine"
	"trading-bot/types"
)

type SMACrossStrategy struct {
	SMAPeriod       int //Hur lång period vi ska kolla efter SMA
	VolumePeriod    int //HUr lång period vi ska kolla efter volym
	VolumeThreshold float64
	SwingLookback   int //Vart vi ska kolla efter swing low/high
	RiskReward      float64 //Hur mycket vi vill riskera
	nyZone          *time.Location
}

// Första SMA stretegi
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

// Returnerar namnet
func (s *SMACrossStrategy) Name() string {
	return "SMA_Cross_Volume"
}

// Kollar att vi har tillräckligt med data
func (s *SMACrossStrategy) OnBar(bar types.Bar, history []types.Bar) types.Signal {
	if len(history) < s.SMAPeriod+1 {
		return types.Signal{Action: "HOLD", Reason: "insufficient data"}
	} 

	//Konverterar från int64 till time.Time
	barTime := time.Unix(bar.Timestamp, 0)
	nyTime := barTime.In(s.nyZone)
	//Kollar att tiden stämmer 
	if !isEarlySession(nyTime) {
		return types.Signal{Action: "HOLD", Reason: "outside early session"}
	}

	//Kollar SMA under olika bars för att se när den korsat
	currentSMA := engine.SMA(history, s.SMAPeriod)
	prevBars := history[:len(history)-1]
	prevSMA := engine.SMA(prevBars, s.SMAPeriod)

	currentBar := history[len(history)-1]
	prevBar := history[len(history)-2]
	//Räknar ut genomsnittsvolymen
	avgVol := engine.AvgVolume(history[:len(history)-1], s.VolumePeriod)
	if float64(currentBar.Volume) < avgVol*s.VolumeThreshold {
		return types.Signal{Action: "HOLD", Reason: "volume too low"}
	}

	//Köp long
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

	//Kort trade
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
