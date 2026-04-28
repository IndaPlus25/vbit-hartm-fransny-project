// types.go
//All strategy files should use this package
package strategies

import "time"
//Extract data from all bars
type Bar struct {
    Time   time.Time 
    Open   float64 //Opening price 
    High   float64 //Highest price
    Low    float64 // Lowest price
    Close  float64 //Closing price
    Volume int64 //Trading volume
}
//Returns after analyzing a bar
type Signal struct {
    Action    string  // "BUY", "SELL", "HOLD"
    Size      float64 //1.0 = 100%
    StopLoss  float64 
    Target    float64
    Reason    string //SMA crossing
}

type Strategy interface {
    Name() string
    OnBar(bar Bar, history []Bar) Signal
}

// Är vi inom första timmarna efter market open? (9:30-12:00 ET)
func isEarlySession(t time.Time) bool {
    hour := t.Hour()
    minute := t.Minute()
    marketMinutes := hour*60 + minute
    // 9:30 = 570 min, 12:00 = 720 min
    return marketMinutes >= 570 && marketMinutes <= 720
}