package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"trading-bot/types"
)

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
