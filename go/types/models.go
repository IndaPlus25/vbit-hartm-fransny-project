package types

type HistoricalTick struct {
	Timestamp int64   `parquet:"timestamp"`
	Price     float64 `parquet:"price"`
	Volume    int64   `parquet:"volume"`
}

type Trade struct {
	Timestamp  int64   `parquet:"timestamp"`
	Symbol     string  `parquet:"symbol"`
	Action     string  `parquet:"action"`
	Price      float64 `parquet:"price"`
	ProfitLoss float64 `parquet:"profit_loss"`
}

type MarketData struct {
	Ticker string  `json:"ticker"`
	Price  float64 `json:"price"`
	SMA    float64 `json:"sma"`
}
