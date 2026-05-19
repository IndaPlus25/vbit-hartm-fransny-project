import pandas as pd
import plotly.graph_objects as go
import os
import platform
import shutil
import subprocess
import webbrowser


def open_in_browser(path: str) -> None:
    """Öppna en lokal fil i standard-webbläsaren. Hanterar WSL (där `gio open`
    saknar file-associations) genom att gå via explorer.exe."""
    abs_path = os.path.abspath(path)
    is_wsl = "microsoft" in platform.uname().release.lower()
    if is_wsl and shutil.which("explorer.exe"):
        try:
            win_path = subprocess.check_output(["wslpath", "-w", abs_path]).decode().strip()
            subprocess.Popen(["explorer.exe", win_path])
            return
        except Exception:
            pass
    webbrowser.open("file://" + abs_path)

def main():
    base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    results_path = os.path.join(base_dir, "go", "output", "results.csv")
    data_dir = os.path.join(base_dir, "data")
    
    if not os.path.exists(results_path):
        print(f"Error: Could not find results file at {results_path}")
        return
        
    df_results = pd.read_csv(results_path)
    if df_results.empty:
        print("No trades found.")
        return

    # Sortera för säkerhets skull så att tickers alltid har samma ordning
    tickers = sorted(df_results['Symbol'].unique())
    
    # Hjälpfunktion för att konvertera tidsstämpel till svensk tid
    to_swe_time = lambda ts: pd.to_datetime(ts, unit='s', utc=True).dt.tz_convert('Europe/Stockholm').dt.tz_localize(None)
    
    fig = go.Figure()
    
    traces_per_ticker = 4
    
    for i, ticker in enumerate(tickers):
        parquet_path = os.path.join(data_dir, f"{ticker}.parquet")
        if not os.path.exists(parquet_path):
            print(f"Warning: Missing data file for {ticker}, skipping.")
            continue
            
        df_ohlcv = pd.read_parquet(parquet_path)
        df_ohlcv['datetime'] = to_swe_time(df_ohlcv['timestamp'])
        
        # Endast den första aktien visas som standard när man laddar sidan
        is_visible = (i == 0)
        
        # 1. Candlestick (Prisgrafen)
        fig.add_trace(go.Candlestick(
            x=df_ohlcv['datetime'],
            open=df_ohlcv['open'],
            high=df_ohlcv['high'],
            low=df_ohlcv['low'],
            close=df_ohlcv['close'],
            name=f"{ticker} Price",
            increasing_line_color='#26A69A', 
            decreasing_line_color='#EF5350',
            visible=is_visible
        ))
        
        # Filtrera trades för denna aktie
        t_df = df_results[df_results['Symbol'] == ticker]
        
        buy_trades = t_df[t_df['Action'] == 'BUY']
        sell_trades = t_df[t_df['Action'] == 'SELL']
        
        # 2. Köpsignaler (Long Entry)
        fig.add_trace(go.Scatter(
            x=to_swe_time(buy_trades['Timestamp']),
            y=buy_trades['EntryPrice'],
            mode='markers',
            marker=dict(symbol='triangle-up', size=14, color='#00E676', line=dict(width=1, color='black')),
            name=f"{ticker} BUY Entry",
            text="<b>" + buy_trades['Strategy'] + "</b><br>PnL: $" + buy_trades['ProfitLoss'].round(2).astype(str),
            hoverinfo='text+x+y',
            visible=is_visible
        ))
        
        # 3. Säljsignaler / Blankning (Short Entry)
        fig.add_trace(go.Scatter(
            x=to_swe_time(sell_trades['Timestamp']),
            y=sell_trades['EntryPrice'],
            mode='markers',
            marker=dict(symbol='triangle-down', size=14, color='#FF5252', line=dict(width=1, color='black')),
            name=f"{ticker} SHORT Entry",
            text="<b>" + sell_trades['Strategy'] + "</b><br>PnL: $" + sell_trades['ProfitLoss'].round(2).astype(str),
            hoverinfo='text+x+y',
            visible=is_visible
        ))
        
        # 4. Exit / Stängda positioner
        fig.add_trace(go.Scatter(
            x=to_swe_time(t_df['CloseTime']),
            y=t_df['ClosePrice'],
            mode='markers',
            marker=dict(symbol='x', size=10, color='#BDBDBD'),
            name=f"{ticker} EXIT",
            text="<b>Exit: " + t_df['Strategy'] + "</b><br>PnL: $" + t_df['ProfitLoss'].round(2).astype(str),
            hoverinfo='text+x+y',
            visible=is_visible
        ))
        
    # Bygg dropdown-menyn för att byta mellan aktier
    buttons = []
    for i, ticker in enumerate(tickers):
        visibility = [False] * (len(tickers) * traces_per_ticker)
        start_idx = i * traces_per_ticker
        
        # Sätt de 4 spåren för denna ticker till synliga
        for j in range(traces_per_ticker):
            if start_idx + j < len(visibility):
                visibility[start_idx + j] = True
            
        buttons.append(dict(
            label=ticker,
            method="update",
            args=[{"visible": visibility},
                  {"title": f"TradingView: {ticker} Trades"}]
        ))
        
    fig.update_layout(
        updatemenus=[dict(
            active=0,
            buttons=buttons,
            x=0.01,
            xanchor="left",
            y=1.05,
            yanchor="top",
            bgcolor="#333333",
            bordercolor="#555555",
            font=dict(color="white")
        )],
        title=dict(text=f"TradingView: {tickers[0]} Trades", font=dict(color='#FFFFFF', size=20)),
        template='plotly_dark',
        plot_bgcolor='#121212',
        paper_bgcolor='#121212',
        xaxis_rangeslider_visible=False, # Dölj den förvald slidern i botten för en renare look
        hovermode='closest'
    )
    
    # Spara och öppna
    out_dir = os.path.join(base_dir, "outline", "images")
    os.makedirs(out_dir, exist_ok=True)
    out_file = os.path.join(out_dir, "tradingview_chart.html")
    
    fig.write_html(out_file)
    print(f"TradingView chart saved to: {out_file}")
    # Öppna i standard-webbläsaren (WSL-säkert)
    open_in_browser(out_file)

if __name__ == "__main__":
    main()
