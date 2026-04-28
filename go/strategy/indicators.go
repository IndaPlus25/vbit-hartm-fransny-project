//Function to track mean price over a period
func SMA(bars []Bar, period int) float64 {
    if len(bars) < period { //Can't calculate SMA without sufficient amount of bars
        return 0
    }
    sum := 0.0
    for i := len(bars) - period; i < len(bars); i++ { //Goes through all bars
        sum += bars[i].Close //Add the closing price
    }
    return sum / float64(period) //Average price durring the period
}
//Average volume durring a period
func AvgVolume(bars []Bar, period int) float64 {
    if len(bars) < period {
        return 0
    }
    sum := int64(0)
    for i := len(bars) - period; i < len(bars); i++ {
        sum += bars[i].Volume
    }
    return float64(sum) / float64(period)
}
//Lowest price durring a period
func SwingLow(bars []Bar, lookback int) float64 {
    if len(bars) < lookback {
        lookback = len(bars)
    }
    lowest := bars[len(bars)-1].Low
    for i := len(bars) - lookback; i < len(bars); i++ {
        if bars[i].Low < lowest {
            lowest = bars[i].Low //Update the lowest value 
        }
    }
    return lowest
}
//Highest price durring a period
func SwingHigh(bars []Bar, lookback int) float64 {
    if len(bars) < lookback {
        lookback = len(bars)
    }
    highest := bars[len(bars)-1].High
    for i := len(bars) - lookback; i < len(bars); i++ {
        if bars[i].High > highest {
            highest = bars[i].High
        }
    }
    return highest
}