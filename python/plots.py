import pandas as pd
import matplotlib.pyplot as plt
import matplotlib.dates as mdates
import os

def main():
    base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    results_path = os.path.join(base_dir, "go", "output", "results.csv")
    
    if not os.path.exists(results_path):
        print(f"Error: Could not find results file at {results_path}")
        return
        
    df = pd.read_csv(results_path)
    
    # Filter closed trades
    closed_trades = df[df['Action'].str.startswith('CLOSE')].copy()
    if closed_trades.empty:
        print("No closed trades to plot.")
        return
        
    # Convert Timestamp to datetime
    closed_trades['Datetime'] = pd.to_datetime(closed_trades['Timestamp'], unit='s')
    
    # Define start and end date for resampling
    start_date = closed_trades['Datetime'].min().floor('D')
    end_date = closed_trades['Datetime'].max().ceil('D')
    daily_idx = pd.date_range(start_date, end_date, freq='D')
    
    # ----- STYLING -----
    plt.style.use('dark_background')
    fig, ax = plt.subplots(figsize=(14, 7), facecolor='#121212')
    ax.set_facecolor('#121212')
    
    colors = ['#00E676', '#29B6F6', '#FFD700', '#AB47BC', '#FF7043']
    strategies = closed_trades['Strategy'].unique()
    
    for i, strat in enumerate(strategies):
        strat_trades = closed_trades[closed_trades['Strategy'] == strat]
        
        # Process Portfolio for this strategy
        strat_pnl = strat_trades.groupby('Datetime')['ProfitLoss'].sum().reset_index()
        strat_pnl = strat_pnl.set_index('Datetime')
        strat_pnl['Cumulative_PnL'] = strat_pnl['ProfitLoss'].cumsum()
        
        # Resample and fill missing days
        strat_daily = strat_pnl['Cumulative_PnL'].resample('D').last()
        strat_daily = strat_daily.reindex(daily_idx).ffill().fillna(0)
        
        color = colors[i % len(colors)]
        
        # Plot strategy performance
        ax.plot(strat_daily.index, strat_daily.values, 
                label=f'Strategy: {strat}', color=color, linewidth=3, zorder=5)
        
        # Fill area under curve
        ax.fill_between(strat_daily.index, strat_daily.values, 0, 
                        where=(strat_daily.values >= 0), 
                        color=color, alpha=0.1, interpolate=True, zorder=4)
        ax.fill_between(strat_daily.index, strat_daily.values, 0, 
                        where=(strat_daily.values < 0), 
                        color=color, alpha=0.05, interpolate=True, zorder=4)
        
    # Formatting
    ax.set_title('STRATEGY COMPARISON: Equity Curve (Daily)', fontsize=20, fontweight='bold', color='#FFFFFF', pad=20)
    ax.set_xlabel('Date', fontsize=14, color='#B3B3B3', labelpad=10)
    ax.set_ylabel('Cumulative Profit/Loss ($)', fontsize=14, color='#B3B3B3', labelpad=10)
    
    # Grid and Spines
    ax.grid(True, color='#2A2A2A', linestyle='--', linewidth=1, alpha=0.7)
    for spine in ['top', 'right']:
        ax.spines[spine].set_visible(False)
    for spine in ['bottom', 'left']:
        ax.spines[spine].set_color('#404040')
        
    ax.tick_params(colors='#B3B3B3', labelsize=11)
    ax.xaxis.set_major_formatter(mdates.DateFormatter('%b %d'))
    plt.xticks(rotation=45)
    
    leg = ax.legend(loc='upper left', frameon=True, facecolor='#1E1E1E', edgecolor='#333333', fontsize=12)
    for text in leg.get_texts():
        text.set_color('#E0E0E0')
        
    plt.tight_layout()
    
    out_dir = os.path.join(base_dir, "outline", "images")
    os.makedirs(out_dir, exist_ok=True)
    out_file = os.path.join(out_dir, "equity_curve.png")
    
    plt.savefig(out_file, dpi=300, bbox_inches='tight', facecolor='#121212')
    print(f"Plot saved successfully to: {out_file}")

if __name__ == "__main__":
    main()
