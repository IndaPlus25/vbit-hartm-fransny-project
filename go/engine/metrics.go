package engine

import (
	"net/http"
	"trading-bot/types"
)

// calculateSMA summerar en slice av priser och dividerar med längden (n).
func calculateSMA(prices []float64) float64 {
	// Implementering
	return 0.0
}

// calculateAverageVolume summerar volymen för att identifiera om
// den aktuella stapeln är över medel (Strategi 1).
func calculateAverageVolume(volumes []int64) float64 {
	// Implementering
	return 0.0
}

// calculateSwingHigh letar iterativt upp det högsta värdet i en array.
func calculateSwingHigh(highs []float64) float64 {
	// Implementering
	return 0.0
}

// calculateSwingLow letar iterativt upp det lägsta värdet i en array.
func calculateSwingLow(lows []float64) float64 {
	// Implementering
	return 0.0
}

// isBullishFVG utvärderar prisgapet mellan det första och tredje ljuset.
func isBullishFVG(c1, c2, c3 HistoricalTick) bool {
	// Implementering: return c3.Low > c1.High
	return false
}

// isBearishFVG utvärderar motsatsen för korta positioner.
func isBearishFVG(c1, c2, c3 HistoricalTick) bool {
	// Implementering: return c3.High < c1.Low
	return false
}

// isBullish returnerar sant om stängningspriset är högre än öppningspriset.
func isBullish(open float64, close float64) bool {
	// Implementering: return close > open
	return false
}

// isBearish returnerar sant om stängningspriset är lägre än öppningspriset.
func isBearish(open float64, close float64) bool {
	// Implementering: return close < open
	return false
}

// isBearishLiquiditySweep verifierar om priset brutit zonen (High > zoneLevel)
// men stängt under den (Close < zoneLevel).
func isBearishLiquiditySweep(high float64, close float64, zoneLevel float64) bool {
	// Implementering
	return false
}

// aggregateTo15Min komprimerar 1-minutsdata till 15-minuters (OHLC).
// Formellt: Open = p_1, High = max(p_1...p_15), Low = min(p_1...p_15), Close = p_15.
func aggregateTo15Min(ticks []HistoricalTick) []HistoricalTick {
	// Implementering
	return nil
}
