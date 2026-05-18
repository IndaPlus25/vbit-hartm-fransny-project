package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"trading-bot/data"
	"trading-bot/engine"
	strategies "trading-bot/strategy" // Import path is the folder name
	"trading-bot/types"
)

func main() {
	fmt.Println("Starting the backtesting engine...")

	// 1. Create a list to store all trades we will make
	var allTrades []types.Trade

	// 2. Instantiate the strategies we want to test
	strats := []engine.Strategy{
		strategies.NewSMACrossStrategy(),
		strategies.NewFVGStrategy(),
		strategies.NewLiquiditySweepStrategy(),
	}

	// 3. Find all downloaded .parquet files in the data directory
	dataDir := filepath.Join("..", "data")
	files, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error reading the data directory: %v\n", err)
		return
	}

	for _, file := range files {
		// Check if the file is a .parquet file
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".parquet") {
			ticker := strings.TrimSuffix(file.Name(), ".parquet")
			filePath := filepath.Join(dataDir, file.Name())

			fmt.Printf("Loading data for %s...\n", ticker)

			// Load the data
			bars, err := data.LoadData(filePath)
			if err != nil {
				fmt.Printf("Could not load data for %s: %v\n", ticker, err)
				continue
			}

			var wg sync.WaitGroup
			var mu sync.Mutex

			// Run the backtest for each strategy
			for _, strat := range strats {
				wg.Add(1)
				go func(s engine.Strategy) {
					defer wg.Done()
					trades := engine.RunBacktest(ticker, bars, s)
					mu.Lock()
					allTrades = append(allTrades, trades...)
					mu.Unlock()
				}(strat)
				fmt.Printf("Running backtest for %s with strategy: %s\n", ticker, strat.Name())
			}
			wg.Wait()
		}
	}

	// 4. Output the results to a CSV file for analysis in Python
	fmt.Printf("\nTotal number of generated trades: %d\n", len(allTrades))
	fmt.Println("Saving results to CSV...")

	err = saveTradesToCSV(allTrades, filepath.Join("output", "results.csv"))
	if err != nil {
		fmt.Printf("Could not save CSV: %v\n", err)
	} else {
		fmt.Println("Backtest complete! Results are available in go/output/results.csv")

		fmt.Println("Generating plot...")

		// Find the correct python command (python or py)
		pythonCmd := "python"
		if _, err := exec.LookPath("python"); err != nil {
			pythonCmd = "py"
		}

		cmd := exec.Command(pythonCmd, filepath.Join("..", "python", "plots.py"))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error generating plot: %v\n", err)
		}
	}
}

// saveTradesToCSV takes a list of Trades and writes it to a CSV file
func saveTradesToCSV(trades []types.Trade, filename string) error {
	// Ensure the output directory exists
	err := os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		return err
	}

	// Create the file itself
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the headers - these must exactly match what the Python script expects
	header := []string{
		"Timestamp",
		"Symbol",
		"Action",
		"Price",
		"ProfitLoss",
		"CloseTime",
		"EntryPrice",
		"ClosePrice",
		"Strategy",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write all data row by row
	for _, t := range trades {
		// Convert numbers to text (string)
		tsStr := strconv.FormatInt(t.Timestamp, 10)
		priceStr := strconv.FormatFloat(t.Price, 'f', 2, 64)
		pnlStr := strconv.FormatFloat(t.ProfitLoss, 'f', 2, 64)
		closeTimeStr := strconv.FormatInt(t.CloseTime, 10)
		entryPriceStr := strconv.FormatFloat(t.EntryPrice, 'f', 2, 64)
		closePriceStr := strconv.FormatFloat(t.ClosePrice, 'f', 2, 64)

		row := []string{
			tsStr,
			t.Symbol,
			t.Action,
			priceStr,
			pnlStr,
			closeTimeStr,
			entryPriceStr,
			closePriceStr,
			t.Strategy,
		}

		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}
