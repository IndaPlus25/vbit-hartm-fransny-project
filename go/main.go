package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"trading-bot/data"
	"trading-bot/engine"
	"trading-bot/network"
	"trading-bot/types"
)

func main() {
	var mu sync.Mutex
	var allTrades []Trade
	var wg sync.WaitGroup

	tickers := []string{} //NEED A SLICE WITH STOCKS

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	allHistoricalData := loadData("INSERT_OWN_FILE") //sätt in rätt filnamn

	for ticker, tickerData := range allHistoricalData {
		wg.Add(1)

		go func(t string, data []HistoricalTick) {
			defer wg.Done()

			stockTrades := simulateTicker(t, httpClient, data)

			mu.Lock()
			allTrades = append(allTrades, stockTrades...)
			mu.Unlock()
		}(ticker, tickerData)
	}
	wg.Wait()

	err := saveToParquet("result.parquet", allTrades)
	if err != nil {
		panic(err)
	}
}
