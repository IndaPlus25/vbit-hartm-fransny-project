package main

import (
	"fmt"
	"os"

	"github.com/segment/parquet-go"
)

type Trade struct {
	Timestamp  int64   `parquet:"timestamp"`
	Symbol     string  `parquet:"symbol"`
	Action     string  `parquet:"action"`
	Price      float64 `parquet:"price"`
	ProfitLoss float64 `parquet:"profit_loss"`
}

func main() {
	signalChan := make(chan TradeSignal)

	//fix go routines
	go runStrategy()
	go runStrategy()
	go runStrategy()

	for signal := range signalChan {
		sendToPython(signal)
	}
}

func sendToPython(s TradeSignal) {
	//implement IPC logic: HTTP here

	fmt.Printf("Pushing to Python: [%s] %s %s\n", s.StrategyID, s.Action, s.Ticker)
}

func runStrategy(name string, ticker string, c chan TradeSignal) {
	for {
		//implement strategies

	}
}

// Help function
// Convert CVS to parquet files (more effective to read)
func saveToParquet(filename string, trades []Trade) error {

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
