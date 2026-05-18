import pandas as pd
import plotly.graph_objects as go
import os
import webbrowser

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
    
    fig = go.Figure()
    
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
        
        # Add a trace for this strategy
        fig.add_trace(go.Scatter(
            x=strat_daily.index, 
            y=strat_daily.values,
            mode='lines',
            name=f'{strat} ({len(strat_trades)} trades)',
            line=dict(color=color, width=3),
            fill='tozeroy',  # Fills the area to zero
            fillcolor=f'rgba{tuple(int(color.lstrip("#")[i:i+2], 16) for i in (0, 2, 4)) + (0.1,)}' # Transparent fill
        ))

    # Update layout to be dark and stylish
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
    
    # Save to HTML file
    out_dir = os.path.join(base_dir, "outline", "images")
    os.makedirs(out_dir, exist_ok=True)
    out_file = os.path.join(out_dir, "equity_curve_interactive.html")
    
    fig.write_html(out_file)
    print(f"Interactive plot saved to: {out_file}")
    
    # Open in default browser
    webbrowser.open('file://' + os.path.abspath(out_file))

if __name__ == "__main__":
    main()
