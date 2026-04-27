package data

import (
	"github.com/segmentio/parquet-go"
	"os"
	"trading-bot/types"
)

func LoadData(filename string) map[string][]HistoricalTick {
	// Här ska Parquet-läsningen för indatan ske senare
	return make(map[string][]HistoricalTick)
}

// Help function that converts CVS to parquet files (more effective to read)
func SaveToParquet(filename string, trades []Trade) error {

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	//Creates new writer only parsing variables of type Trade.
	//writer has parquet instruction (schema-based format)
	writer := parquet.NewGenericWriter[Trade](file)
	defer writer.Close()

	_, err = writer.Write(trades)
	if err != nil {
		panic(err)
	}
	return nil
}
