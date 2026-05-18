import pandas as pd
import os

def main():
    base_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    results_path = os.path.join(base_dir, "go", "output", "results.csv")
    
    if not os.path.exists(results_path):
        print(f"Error: Could not find results file at {results_path}")
        return
        
    df = pd.read_csv(results_path)
    if df.empty:
        print("No trades found in the results file.")
        return
        
    # Go-motorn exporterar nu bara stängda trades, så alla rader är giltiga
    closed_trades = df.copy()
    
    if closed_trades.empty:
        print("No closed trades found.")
        return
        
    # Analysera per strategi
    print("=" * 55)
    print("MULTI-STRATEGY BACKTEST RESULTS")
    print("=" * 55)
    
    strategies = closed_trades['Strategy'].unique()
    for strat in strategies:
        strat_trades = closed_trades[closed_trades['Strategy'] == strat].copy()
        
        total_trades = len(strat_trades)
        winning_trades = len(strat_trades[strat_trades['ProfitLoss'] > 0])
        win_rate = (winning_trades / total_trades) * 100 if total_trades > 0 else 0
        total_pnl = strat_trades['ProfitLoss'].sum()
        
        # Räkna ut Max Drawdown
        strat_trades = strat_trades.sort_values('Timestamp')
        cumulative_pnl = strat_trades['ProfitLoss'].cumsum()
        peak = cumulative_pnl.expanding(min_periods=1).max()
        drawdown = peak - cumulative_pnl
        max_drawdown = drawdown.max()
        
        print(f"\n--- Strategy: {strat} ---")
        print(f"Total Closed Trades: {total_trades}")
        print(f"Win Rate:            {win_rate:.2f}%")
        print(f"Total Profit/Loss:   {total_pnl:.2f}")
        print(f"Max Drawdown:        {max_drawdown:.2f}")

if __name__ == "__main__":
    main()
