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
    
    if not os.path.exists(results_path):
        print(f"Error: Could not find results file at {results_path}")
        return
        
    df = pd.read_csv(results_path)
    
    # Go-motorn exporterar bara stängda trades
    closed_trades = df.copy()
    if closed_trades.empty:
        print("No closed trades to plot.")
        return
        
    # Konvertera tidsstämpel till svensk tid
    closed_trades['Datetime'] = pd.to_datetime(closed_trades['Timestamp'], unit='s', utc=True).dt.tz_convert('Europe/Stockholm').dt.tz_localize(None)
    
    # Definiera start och slut för datum
    start_date = closed_trades['Datetime'].min().floor('D')
    plot_start_date = start_date - pd.Timedelta(days=1)
    end_date = closed_trades['Datetime'].max().ceil('D')
    daily_idx = pd.date_range(plot_start_date, end_date, freq='D')
    
    fig = go.Figure()
    
    colors = ['#00E676', '#29B6F6', '#FFD700', '#AB47BC', '#FF7043']
    strategies = closed_trades['Strategy'].unique()
    
    for i, strat in enumerate(strategies):
        strat_trades = closed_trades[closed_trades['Strategy'] == strat]
        
        # Räkna ut portföljens utveckling för denna strategi
        strat_pnl = strat_trades.groupby('Datetime')['ProfitLoss'].sum().reset_index()
        strat_pnl = strat_pnl.set_index('Datetime')
        strat_pnl['Cumulative_PnL'] = 10000.0 + strat_pnl['ProfitLoss'].cumsum()
        
        # Börja alla grafer på 10 000
        strat_pnl.loc[plot_start_date, 'Cumulative_PnL'] = 10000.0
        strat_pnl = strat_pnl.sort_index()
        
        # Fyll i saknade dagar
        strat_daily = strat_pnl['Cumulative_PnL'].resample('D').last()
        strat_daily = strat_daily.reindex(daily_idx).ffill().fillna(10000.0)
        
        color = colors[i % len(colors)]
        
        # Rita ut strategin
        fig.add_trace(go.Scatter(
            x=strat_daily.index, 
            y=strat_daily.values,
            mode='lines',
            name=f'{strat} ({len(strat_trades)} trades)',
            line=dict(color=color, width=3)
        ))

    # Designa layouten (mörkt tema)
    fig.update_layout(
        title={
            'text': 'STRATEGY COMPARISON: Equity Curve (Daily)',
            'y':0.95,
            'x':0.5,
            'xanchor': 'center',
            'yanchor': 'top',
            'font': dict(size=24, color='#FFFFFF', family="Arial, sans-serif")
        },
        xaxis_title='Date',
        yaxis_title='Cumulative Profit/Loss ($)',
        template='plotly_dark',
        plot_bgcolor='#121212',
        paper_bgcolor='#121212',
        font=dict(color='#B3B3B3'),
        xaxis=dict(
            showgrid=True,
            gridcolor='#2A2A2A',
            gridwidth=1,
            zeroline=False
        ),
        yaxis=dict(
            showgrid=True,
            gridcolor='#2A2A2A',
            gridwidth=1,
            zeroline=True,
            zerolinecolor='#404040',
            zerolinewidth=2
        ),
        hovermode='x unified',
        legend=dict(
            yanchor="top",
            y=0.99,
            xanchor="left",
            x=0.01,
            bgcolor='rgba(30, 30, 30, 0.8)',
            bordercolor='#333333',
            borderwidth=1
        )
    )
    
    # Spara grafen som HTML
    out_dir = os.path.join(base_dir, "outline", "images")
    os.makedirs(out_dir, exist_ok=True)
    out_file = os.path.join(out_dir, "equity_curve_interactive.html")
    
    fig.write_html(out_file)
    print(f"Interactive plot saved to: {out_file}")
    
    # Öppna i standard-webbläsaren (WSL-säkert)
    open_in_browser(out_file)

if __name__ == "__main__":
    main()
