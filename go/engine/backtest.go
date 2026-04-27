package engine

import (
	"net/http"
	"trading-bot/network"
	"trading-bot/types"
)

func simulateTicker(ticker string, client *http.Client, data []HistoricalTick) []Trade {
	var tradesForTicker []Trade

	//read data
	//loop over time
	//MATH functions
	//IPC: build JSON and do https.Post to Python
	//Register: What did python send back after strats??

	//SMA math

	url := "http://localhost:8000/signal"
	signal, err := sendToPython(client, url, aktuellData)
	if err != nil {
		fmt.Println("Fel vid anrop:", err)
	}

	return tradesForTicker
}
