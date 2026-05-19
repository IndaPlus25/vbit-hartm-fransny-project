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

	// Skapar lista över aktuella trades
	var allTrades []types.Trade

	// Instantiate the strategies we want to test
	strats := []engine.Strategy{
		strategies.NewSMACrossStrategy(),
		strategies.NewFVGStrategy(),
		strategies.NewLiquiditySweepStrategy(),
	}

	// Hitta alla nedladdade parquetfiler i data.go mappen
	dataDir := filepath.Join("..", "data")
	files, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error reading the data directory: %v\n", err)
		return
	}

	for _, file := range files {
		// Kollar om filen är en .parquet fil
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".parquet") {
			ticker := strings.TrimSuffix(file.Name(), ".parquet")
			filePath := filepath.Join(dataDir, file.Name())

			fmt.Printf("Loading data for %s...\n", ticker)

			// Ladda data
			bars, err := data.LoadData(filePath)
			if err != nil {
				fmt.Printf("Could not load data for %s: %v\n", ticker, err)
				continue
			}

			var wg sync.WaitGroup
			var mu sync.Mutex

			// Kör backtest
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

	// Output som CSV fil för att hantera i python
	fmt.Printf("\nTotal number of generated trades: %d\n", len(allTrades))
	fmt.Println("Saving results to CSV...")

	err = saveTradesToCSV(allTrades, filepath.Join("output", "results.csv"))
	if err != nil {
		fmt.Printf("Could not save CSV: %v\n", err)
	} else {
		fmt.Println("Backtest complete! Results are available in go/output/results.csv")

		fmt.Println("Generating plot...")

		// 1. Hitta rätt python-kommando: föredra projektets .venv om den finns
		// (söker uppåt i katalogträdet upp till 5 nivåer), annars python3/python/py i PATH.
		pythonCmd := ""
		venvSubpaths := []string{
			filepath.Join(".venv", "bin", "python3"),
			filepath.Join(".venv", "bin", "python"),
			filepath.Join(".venv", "Scripts", "python.exe"),
		}
		base := ".."
		for i := 0; i < 5 && pythonCmd == ""; i++ {
			for _, sub := range venvSubpaths {
				candidate := filepath.Join(base, sub)
				if _, err := os.Stat(candidate); err == nil {
					pythonCmd = candidate
					break
				}
			}
			base = filepath.Join(base, "..")
		}
		if pythonCmd == "" {
			for _, candidate := range []string{"python3", "python", "py"} {
				if _, err := exec.LookPath(candidate); err == nil {
					pythonCmd = candidate
					break
				}
			}
		}
		if pythonCmd == "" {
			fmt.Println("Error: hittade ingen python (kollade .venv uppåt i trädet och PATH)")
			return
		}

		// 2. Kör det första skriptet: plot_growth.py
		fmt.Println("-> Running plot_growth.py...")
		cmdGrowth := exec.Command(pythonCmd, filepath.Join("..", "python", "plot_growth.py"))
		cmdGrowth.Stdout = os.Stdout
		cmdGrowth.Stderr = os.Stderr
		if err := cmdGrowth.Run(); err != nil {
			fmt.Printf("Error generating growth plot: %v\n", err)
		}

		// 3. Kör det andra skriptet: plot_trades.pya
		fmt.Println("-> Running plot_trades.py...")
		cmdTrades := exec.Command(pythonCmd, filepath.Join("..", "python", "plot_trades.py"))
		cmdTrades.Stdout = os.Stdout
		cmdTrades.Stderr = os.Stderr
		if err := cmdTrades.Run(); err != nil {
			fmt.Printf("Error generating trades plot: %v\n", err)
		}
	}
}

// saveTradesToCSV tar en lista av trades och skriver om till en CSV fil.
func saveTradesToCSV(trades []types.Trade, filename string) error {
	// Säkerställer att en output mapp finns
	err := os.MkdirAll(filepath.Dir(filename), os.ModePerm)
	if err != nil {
		return err
	}

	// Skapa själva filen
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Skriv rubrikerna som måste matcha i python
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

	for _, t := range trades {
		// Gör nummer till text
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

