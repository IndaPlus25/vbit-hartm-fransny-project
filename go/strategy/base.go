// types.go
//All strategy files should use this package

package strategies

import (
	"time"
	"trading-bot/types"
)

type Strategy interface {
	Name() string
	OnBar(bar types.Bar, history []types.Bar) types.Signal
}

// Är vi inom första timmarna efter market open? (9:30-12:00 ET)
func isEarlySession(t time.Time) bool {
	hour := t.Hour()
	minute := t.Minute()
	marketMinutes := hour*60 + minute
	// 9:30 = 570 min, 12:00 = 720 min
	return marketMinutes >= 570 && marketMinutes <= 720
}
