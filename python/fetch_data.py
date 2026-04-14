import yfinance as yf
import pandas as pd
import os
import argparse
import sys

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--tickers", nargs='+', default=["NVDA", "AAPL", "MSFT"])
    parser.add_argument("--period", type=str, default="60d")
    parser.add_argument("--interval", type=str, default="5m")
    args = parser.parse_args()

    tickers = args.tickers
    period = args.period
    interval = args.interval

    print(f"Fetching historical stock data for: {', '.join(tickers)} using period: {period} and interval: {interval}...")
    
    try:
        # Fetch data from Yahoo Finance
        data = yf.download(tickers, period=period, interval=interval)
        
        if data.empty:
            print("Warning: No data was fetched from Yahoo Finance. Please check the tickers or your internet connection.")
            return

        # We focus on the Adjusted Close prices if available (accounts for splits/dividends)
        if 'Adj Close' in data:
            prices = data['Adj Close']
        elif 'Close' in data:
            prices = data['Close']
        else:
            prices = data

        # Handle missing values by forward-filling the prices
        # and then backward-filling if the first days are NaN
        prices = prices.ffill().bfill()
            
        # Ensure the data directory exists
        base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        target_dir = os.path.join(base_dir, "data")
        os.makedirs(target_dir, exist_ok=True)
        
        csv_path = os.path.join(target_dir, "prices.csv")
        
        # Save to prices.csv
        try:
            prices.to_csv(csv_path)
            print(f"Data fetch complete. Prices have been saved to: {csv_path}")
        except PermissionError:
            print(f"Error: Could not write to {csv_path}. Is the file open in Excel or another program?")
            
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

if __name__ == "__main__":
    main()
