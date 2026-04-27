package main

import (
	"fmt"
	"os"
	"sync"
	"io"
	"net/http"
	"time"
	"encoding/json"
	"bytes"
	
	"github.com/segment/parquet-go"
)

type Trade struct {
	Timestamp  int64   `parquet:"timestamp"`
	Symbol     string  `parquet:"symbol"`
	Action     string  `parquet:"action"`
	Price      float64 `parquet:"price"`
	ProfitLoss float64 `parquet:"profit_loss"`
}

type MarketData struct {
	Ticker string `json:"ticker"`
	Price float64 `json:"price"`
	SMA float64 `json:"sma"`
}

func main() {
	var mu sync.Mutex
	var allTrades []Trade
	var wg sync.WaitGroup

	tickerss := []string{} //NEED A SLICE WITH STOCKS

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, ticker := range tickers {
		wg.Add(1)

		go func(t string) {
			defer wg.Done()
			stockTrades := simulateTicker(t, httpClient)
			mu.Lock()
			allTrades = append(allTrades, stockTrades...)
			mu.Unlock()
		}(ticker)
	}
	wg.Wait()

	err := saveToParquet("result.parquet", allTrades)
	if err != nil {
		panic(err)
	}
}

func simulateTicker(ticker string) []Trade {
	var tradesForTicker []Trade

	//read data
	//loop over time
	//MATH: calculate SMA RSI (parameters) to a point i
	//IPC: build JSON and do https.Post to Python
	//Register: What did python send back after strats??
	
	url := "http://localhost:8000/signal"
	signal, err := sendToPython(client, url, aktuellData)
	if err != nil {
		fmt.Println("Fel vid anrop:", err)
	}
	

	return tradesForTicker
}

func sendToPython(client *http.Client, url string, data MarketData) (string, error) {
	//rewrites data to JSON fomrat
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("Could not code JSON: %v", err)
	}

	//creates the "bridge" between Go and Python
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("Could not create request: %v", err)
	}

	//tells Python that the data is JSON format
	req.Header.Set("Content-Type", "application/json")

	//Sends TCP package over local host and goroutine blocks until python gives response
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("network error towards python %v", err)
	}
	defer resp.Body.Close()

	//reades the Python respond, stores in the array bodyBytes and casts to a Go string  
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Could not read python-post: %v", err)
	}

	//returns python analysis result after strategies has been implemented: BUY, SELL, HOLD
	return string(bodyBytes), nil
}

// Help function that converts CVS to parquet files (more effective to read)
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
