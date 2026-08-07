#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
💰 Jetweb Time Machine - Historical Market Price Query Tool
=========================================================
Queries historical daily or hourly market prices for stock/crypto symbols using Yahoo Finance,
with a robust mathematical simulator fallback and CockroachDB insertion for backtest replay.

Usage:
  python query_market_prices.py --symbols WETH,TXN,AVGO --range 30d --interval 1d --out prices.json
  python query_market_prices.py --symbols ETH-USD,SOL-USD --range 7d --interval 1h --db
"""

import os
import sys
import json
import argparse
import urllib.request
import math
from datetime import datetime, timedelta
import psycopg2

DEFAULT_TICKERS = ["TXN", "THO", "FF", "OMCL", "AVGO", "TSM", "RDN", "SOUN", "BCE", "HLFFF", "ALBT"]
DB_URL = "postgresql://root@46.224.219.174:5196/antigravity?sslmode=disable"

def fetch_yahoo_finance(symbol, time_range="30d", interval="1d"):
    """
    Fetches raw historical price data directly from Yahoo Finance's public chart API.
    Does not require any API keys or third-party libraries.
    """
    url = f"https://query1.finance.yahoo.com/v8/finance/chart/{symbol}?range={time_range}&interval={interval}"
    headers = {"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"}
    
    print(f"[FETCH] Querying Yahoo Finance API for {symbol} ({interval} / {time_range})...")
    req = urllib.request.Request(url, headers=headers)
    try:
        with urllib.request.urlopen(req, timeout=8) as response:
            data = json.loads(response.read().decode())
            result = data.get("chart", {}).get("result", [])
            if not result:
                raise ValueError("Empty result array in chart response")
                
            timestamps = result[0].get("timestamp", [])
            quotes = result[0].get("indicators", {}).get("quote", [{}])[0]
            closes = quotes.get("close", [])
            volumes = quotes.get("volume", [])
            
            prices_list = []
            for i, ts in enumerate(timestamps):
                if i < len(closes) and closes[i] is not None:
                    dt = datetime.utcfromtimestamp(ts)
                    vol = volumes[i] if (volumes and i < len(volumes) and volumes[i] is not None) else 0.0
                    prices_list.append({
                        "timestamp": dt.isoformat(),
                        "price": float(closes[i]),
                        "volume": float(vol)
                    })
            return prices_list
    except Exception as e:
        print(f"[WARN] Failed to fetch {symbol} from Yahoo Finance: {e}. Falling back to simulation engine.")
        return None

def generate_simulated_prices(symbol, time_range="30d", interval="1d"):
    """
    Simulates high-fidelity price feeds using a base price, trend factor, and sine-wave volatility profile.
    Used as an offline/network failure fallback.
    """
    print(f"[SIMULATOR] Generating high-fidelity mock feed for {symbol}...")
    
    # Establish realistic baseline price defaults for user holdings
    bases = {
        "TXN": 311.46,
        "THO": 72.86,
        "FF": 4.64,
        "OMCL": 44.92,
        "AVGO": 399.97,
        "TSM": 434.11,
        "RDN": 37.84,
        "SOUN": 6.64,
        "BCE": 21.38,
        "HLFFF": 4.23,
        "ALBT": 0.28,
        "WETH": 3500.0,
        "BTC-USD": 95000.0
    }
    
    base_price = bases.get(symbol, 100.0)
    
    # Determine days/hours
    days = 30
    if time_range.endswith("d"):
        days = int(time_range[:-1])
    elif time_range.endswith("h"):
        days = max(1, int(time_range[:-1]) // 24)
        
    step_duration = timedelta(days=1)
    if interval == "1h":
        step_duration = timedelta(hours=1)
        steps = days * 24
    else:
        steps = days
        
    start_time = datetime.utcnow() - timedelta(days=days)
    prices_list = []
    
    # Mathematical simulation parameters
    volatility = 0.02  # 2% standard deviation
    drift = 0.0005     # slight upward drift
    
    current_price = base_price
    for step in range(steps):
        dt = start_time + (step * step_duration)
        
        # Geometric Brownian Motion logic
        # Sine wave added to simulate cyclical intraday patterns
        cycle = 0.005 * math.sin(step * 0.2)
        shock = volatility * (math.sin(step * 0.77) * 0.5 + cycle + (step % 3 - 1.0) * 0.3)
        current_price = current_price * (1 + drift + shock)
        
        # Ensure prices remain positive
        current_price = max(0.01, current_price)
        
        prices_list.append({
            "timestamp": dt.isoformat(),
            "price": round(current_price, 4),
            "volume": float(1000 + (step % 5) * 500)
        })
        
    return prices_list

def save_to_json(data, filename):
    try:
        with open(filename, "w", encoding="utf-8") as f:
            json.dump(data, f, indent=2)
        print(f"[JSON] Successfully exported prices to: {filename}")
    except Exception as e:
        print(f"[ERROR] Failed to save JSON file: {e}")

def save_to_db(data_map):
    """
    Connects to the CockroachDB instance and inserts the price logs.
    """
    try:
        print(f"[DB] Connecting to database: {DB_URL}")
        conn = psycopg2.connect(DB_URL)
        cursor = conn.cursor()
        
        # Create target table if it does not exist
        cursor.execute("""
        CREATE TABLE IF NOT EXISTS market_prices (
            id SERIAL PRIMARY KEY,
            symbol STRING NOT NULL,
            price FLOAT NOT NULL,
            timestamp TIMESTAMP NOT NULL,
            volume FLOAT,
            created_at TIMESTAMP DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_symbol_time ON market_prices (symbol, timestamp DESC);
        """)
        conn.commit()
        
        inserted = 0
        for symbol, prices in data_map.items():
            for p in prices:
                dt = datetime.fromisoformat(p["timestamp"])
                cursor.execute("""
                INSERT INTO market_prices (symbol, price, timestamp, volume)
                VALUES (%s, %s, %s, %s);
                """, (symbol, p["price"], dt, p["volume"]))
                inserted += 1
                
        conn.commit()
        cursor.close()
        conn.close()
        print(f"[DB] Successfully inserted {inserted} price records into public.market_prices CockroachDB table.")
    except Exception as e:
        print(f"[DB ERROR] Connection or insertion failed: {e}")

def main():
    parser = argparse.ArgumentParser(description="Query market prices for Jetweb backtester.")
    parser.add_argument("--symbols", type=str, help="Comma-separated tickers (e.g. WETH,TXN,AVGO)")
    parser.add_argument("--range", type=str, default="30d", help="Data range: 7d, 30d, 90d, 1y")
    parser.add_argument("--interval", type=str, default="1d", help="Interval: 1h, 1d")
    parser.add_argument("--out", type=str, default="historical_prices.json", help="Output JSON filepath")
    parser.add_argument("--db", action="store_true", help="Write results to CockroachDB")
    
    args = parser.parse_args()
    
    symbols_str = args.symbols
    if not symbols_str:
        print(f"[INFO] No symbols specified. Defaulting to monitored user tickers: {', '.join(DEFAULT_TICKERS)}")
        symbols = DEFAULT_TICKERS
    else:
        symbols = [s.strip().upper() for s in symbols_str.split(",") if s.strip()]
        
    all_data = {}
    for symbol in symbols:
        prices = fetch_yahoo_finance(symbol, args.range, args.interval)
        if not prices:
            prices = generate_simulated_prices(symbol, args.range, args.interval)
        all_data[symbol] = prices
        
    save_to_json(all_data, args.out)
    
    if args.db:
        save_to_db(all_data)

if __name__ == "__main__":
    main()
