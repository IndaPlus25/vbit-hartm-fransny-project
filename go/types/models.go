package types

type Trade struct {
	Timestamp  int64   `parquet:"timestamp"`
	Symbol     string  `parquet:"symbol"`
	Action     string  `parquet:"action"`
	Price      float64 `parquet:"price"`
	ProfitLoss float64 `parquet:"profit_loss"`
	CloseTime  int64   `parquet:"close_time"`
	EntryPrice float64 `parquet:"entry_price"`
	ClosePrice float64 `parquet:"close_price"`
	Strategy   string  `parquet:"strategy"`
}

type MarketData struct {
	Ticker string  `json:"ticker"`
	Price  float64 `json:"price"`
	SMA    float64 `json:"sma"`
}

// Extract data from all bars
type Bar struct {
	Timestamp int64   `parquet:"timestamp"`
	Open      float64 `parquet:"open"`   //Opening price
	High      float64 `parquet:"high"`   //Highest price
	Low       float64 `parquet:"low"`    // Lowest price
	Close     float64 `parquet:"close"`  //Closing price
	Volume    int64   `parquet:"volume"` //Trading volume
}

// Returns after analyzing a bar
type Signal struct {
	Action   string  // "BUY", "SELL", "HOLD"
	Size     float64 //1.0 = 100%
	StopLoss float64
	Target   float64
	Reason   string //SMA crossing
}
