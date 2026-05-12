package types

// every file with functions are part of Startegy (fvg, LIQ, cross)
type Strategy interface {
	Name() string
	OnBar(bar Bar, history []Bar) Signal
}
