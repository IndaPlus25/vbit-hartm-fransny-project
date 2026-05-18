import yfinance as yf
import pandas as pd
import os
import argparse
import sys

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--tickers", nargs='+', default=["NVDA", "AAPL", "MSFT"])
    parser.add_argument("--period", type=str, default="1y")
    parser.add_argument("--interval", type=str, default="5m")
    args = parser.parse_args()

    tickers = args.tickers
    period = args.period
    interval = args.interval

    # Yahoo Finance tillåter max 60 dagar för 5-minutersintervall
    if interval == "5m" and period not in ["1d", "5d", "1mo", "60d"]:
        print("Note: yfinance limits 5m data to a maximum of 60 days. Adjusting period to '60d'.")
        period = "60d"

    print(f"Fetching historical stock data for: {', '.join(tickers)} using period: {period} and interval: {interval}...")
    
    try:
        # Hämta data grupperat per aktie
        data = yf.download(tickers, period=period, interval=interval, group_by='ticker')
        
        if data.empty:
            print("Warning: No data was fetched from Yahoo Finance. Please check the tickers or your internet connection.")
            return

        base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        target_dir = os.path.join(base_dir, "data")
        os.makedirs(target_dir, exist_ok=True)
        
        for ticker in tickers:
            # Hantera formatet beroende på antal aktier
            if len(tickers) == 1:
                ticker_data = data
            else:
                ticker_data = data[ticker]
                
            # Rensa tomma rader
            ticker_data = ticker_data.dropna()
            
            # Gör datum till en vanlig kolumn
            df = ticker_data.reset_index()
            datetime_col = df.columns[0]
            
            # Bygg dataframen för Go-motorn
            export_df = pd.DataFrame()
            
            # Unix-tidsstämpel (sekunder)
            export_df['timestamp'] = df[datetime_col].astype('int64') // 10**9 
            
            # Formatera OHLCV-kolumnerna
            export_df['open'] = df['Open'].astype('float64')
            export_df['high'] = df['High'].astype('float64')
            export_df['low'] = df['Low'].astype('float64')
            export_df['close'] = df['Close'].astype('float64') 
            export_df['volume'] = df['Volume'].astype('int64')
            
            parquet_path = os.path.join(target_dir, f"{ticker}.parquet")
            
            try:
                # Spara som parquet
                export_df.to_parquet(parquet_path, engine='pyarrow', index=False)
                print(f"Saved OHLCV data for {ticker} to: {parquet_path}")
            except ImportError:
                print("Error: 'pyarrow' or 'fastparquet' is required to save parquet files. Run 'pip install pyarrow'.")
                return
            except Exception as e:
                print(f"Error saving {ticker}: {e}")
            
    except Exception as e:
        print(f"An unexpected error occurred: {e}")

if __name__ == "__main__":
    main()
