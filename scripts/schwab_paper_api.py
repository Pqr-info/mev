#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
🏦 Charles Schwab Mock Paper Trading API Server
============================================
Launches a local REST API server mimicking Charles Schwab's API,
initializing portfolio state from 'schwab_summary_dom.json' and simulating
market execution using offline prices.

Default Port: 8085
Endpoints:
  GET  /api/v1/accounts/summary
  GET  /api/v1/accounts/positions
  POST /api/v1/orders/place
  GET  /api/v1/orders
  POST /api/v1/orders/cancel
"""

import os
import json
import re
import time
import math
from http.server import HTTPServer, BaseHTTPRequestHandler
import urllib.parse
from datetime import datetime
import psycopg2
import telegram_notifier

PORT = 8085
DB_URL = "postgresql://root@46.224.219.174:5196/antigravity?sslmode=disable"

# Fallback path checking for local state and scraped DOM
CURRENT_BRAIN_DIR = r"C:\Users\theal\.gemini\antigravity\brain\c431ba66-329c-4cff-9571-25470fef9831"
LEGACY_BRAIN_DIR = r"C:\Users\theal\.gemini\antigravity-cli\brain\27cc7b87-864e-4692-aae4-745760bc1eb9"

BRAIN_DIR = CURRENT_BRAIN_DIR if os.path.exists(CURRENT_BRAIN_DIR) else LEGACY_BRAIN_DIR
DOM_PATH = os.path.join(BRAIN_DIR, "scratch", "schwab_summary_dom.json")
STATE_PATH = os.path.join(BRAIN_DIR, "scratch", "paper_portfolio.json")

# Ensure scratch directory exists
os.makedirs(os.path.dirname(DOM_PATH), exist_ok=True)

class PortfolioManager:
    _cached_portfolio = None
    _last_state_mtime = 0
    _last_dom_mtime = 0

    @staticmethod
    def initialize_from_dom():
        """
        Parses schwab_summary_dom.json to bootstrap cash and initial stock positions.
        """
        print("[INIT] Bootstrapping portfolio from Schwab DOM...")
        initial_state = {
            "cash": 114995.70,
            "positions": {},
            "orders": [],
            "history": []
        }
        
        # Safe default positions if DOM parsing fails
        default_positions = {
            "TXN": {"qty": 7.699, "cost_basis": 1519.53, "price": 311.46},
            "THO": {"qty": 25.0, "cost_basis": 2606.75, "price": 72.86},
            "FF": {"qty": 302.0, "cost_basis": 1790.88, "price": 4.64},
            "OMCL": {"qty": 30.0, "cost_basis": 1480.20, "price": 44.92},
            "AVGO": {"qty": 3.0, "cost_basis": 894.87, "price": 399.97},
            "TSM": {"qty": 2.0, "cost_basis": 649.00, "price": 434.11},
            "RDN": {"qty": 20.0, "cost_basis": 672.60, "price": 37.84},
            "SOUN": {"qty": 50.0, "cost_basis": 338.00, "price": 6.64},
            "BCE": {"qty": 10.0, "cost_basis": 257.80, "price": 21.38},
            "HLFFF": {"qty": 30.0, "cost_basis": 365.00, "price": 4.23},
            "ALBT": {"qty": 1.0, "cost_basis": 1.19, "price": 0.28}
        }
        
        if not os.path.exists(DOM_PATH):
            print("[WARN] schwab_summary_dom.json not found. Using hardcoded defaults.")
            initial_state["positions"] = default_positions
            return initial_state
            
        try:
            from bs4 import BeautifulSoup
            with open(DOM_PATH, "r", encoding="utf-8") as f:
                data = json.load(f)
            soup = BeautifulSoup(data.get("dom", ""), "html.parser")
            
            # Find equities table rows
            for tr in soup.find_all("tr"):
                cells = tr.find_all(["td", "th"])
                if len(cells) < 11:
                    continue
                txts = [c.get_text(strip=True) for c in cells]
                
                # Check for standard stock ticker pattern (e.g. TXNTXN...)
                symbol_match = re.match(r"^([A-Z]{1,5})([A-Z]{1,5})", txts[0])
                if symbol_match and symbol_match.group(1) == symbol_match.group(2):
                    ticker = symbol_match.group(1)
                    try:
                        # Extract quantity, price, cost basis
                        qty_str = txts[4].replace("Quantity", "").replace(",", "").strip()
                        qty = float(qty_str) if qty_str else 0.0
                        
                        price_str = txts[5].replace("Price$", "").replace("?", "").replace(",", "").strip()
                        price = float(price_str) if price_str else 0.0
                        
                        cost_str = txts[9].replace("Cost Basis$", "").replace(",", "").strip()
                        cost_basis = float(cost_str) if cost_str else 0.0
                        
                        if qty > 0 and price > 0:
                            initial_state["positions"][ticker] = {
                                "qty": qty,
                                "price": price,
                                "cost_basis": cost_basis
                            }
                    except Exception as e:
                        print(f"[WARN] Error parsing row {txts[0]}: {e}")
                        
            # Extract cash balance from DOM if available
            for tr in soup.find_all("tr"):
                text = tr.get_text()
                if "Cash & Cash Investments" in text:
                    m = re.search(r"\$([0-9,]+\.[0-9]{2})", text)
                    if m:
                        initial_state["cash"] = float(m.group(1).replace(",", ""))
                        break
                        
            print(f"[INIT] Parsed {len(initial_state['positions'])} positions from DOM.")
            if len(initial_state["positions"]) == 0:
                print("[WARN] No positions parsed from DOM. Falling back to default positions.")
                initial_state["positions"] = default_positions
        except Exception as e:
            print(f"[ERROR] Failed to parse DOM: {e}. Falling back to default positions.")
            initial_state["positions"] = default_positions
            
        return initial_state

    @classmethod
    def load_portfolio(cls):
        dom_mtime = 0
        if os.path.exists(DOM_PATH):
            try:
                dom_mtime = os.path.getmtime(DOM_PATH)
            except Exception:
                pass

        state_mtime = 0
        if os.path.exists(STATE_PATH):
            try:
                state_mtime = os.path.getmtime(STATE_PATH)
            except Exception:
                pass

        # Return cached portfolio if neither DOM nor state changed on disk
        if cls._cached_portfolio is not None:
            if dom_mtime <= cls._last_dom_mtime and state_mtime <= cls._last_state_mtime:
                return cls._cached_portfolio

        # Load from state file if present
        if os.path.exists(STATE_PATH):
            try:
                with open(STATE_PATH, "r", encoding="utf-8") as f:
                    cls._cached_portfolio = json.load(f)
                    cls._last_state_mtime = state_mtime
                    cls._last_dom_mtime = dom_mtime
                    return cls._cached_portfolio
            except Exception as e:
                print(f"[ERROR] Loading state failed: {e}")
        
        # Initialize and save initial state
        state = cls.initialize_from_dom()
        cls._cached_portfolio = state
        cls._last_dom_mtime = dom_mtime
        cls.save_portfolio(state)
        return state

    @classmethod
    def save_portfolio(cls, state):
        cls._cached_portfolio = state
        try:
            with open(STATE_PATH, "w", encoding="utf-8") as f:
                json.dump(state, f, indent=2)
            if os.path.exists(STATE_PATH):
                cls._last_state_mtime = os.path.getmtime(STATE_PATH)
        except Exception as e:
            print(f"[ERROR] Saving state failed: {e}")

def get_current_price(symbol):
    """
    Attempts to fetch current price from CockroachDB public.market_prices,
    falling back to mathematical mock volatility calculation.
    """
    try:
        conn = psycopg2.connect(DB_URL)
        cursor = conn.cursor()
        cursor.execute("SELECT price FROM market_prices WHERE symbol = %s ORDER BY timestamp DESC LIMIT 1;", (symbol,))
        row = cursor.fetchone()
        cursor.close()
        conn.close()
        if row:
            return float(row[0])
    except Exception:
        pass
        
    # Sine wave fallback pricing simulator
    base_prices = {"TXN": 311.46, "AVGO": 399.97, "THO": 72.86, "OMCL": 44.92, "TSM": 434.11}
    base = base_prices.get(symbol, 100.0)
    cycle = math.sin(time.time() * 0.05) * 0.02 # +/- 2% swing
    return round(base * (1 + cycle), 2)

class SchwabRequestHandler(BaseHTTPRequestHandler):
    def _send_json(self, status, data):
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(json.dumps(data).encode("utf-8"))

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        self.end_headers()

    def do_GET(self):
        parsed_url = urllib.parse.urlparse(self.path)
        path = parsed_url.path
        
        # Serve local dashboard static file directly at root path
        if path in ["/", "/index.html", "/dashboard", "/orderbook"]:
            html_path = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "dashboard", "schwab_orderbook.html")
            if os.path.exists(html_path):
                self.send_response(200)
                self.send_header("Content-Type", "text/html")
                self.send_header("Access-Control-Allow-Origin", "*")
                self.end_headers()
                try:
                    with open(html_path, "r", encoding="utf-8") as f:
                        self.wfile.write(f.read().encode("utf-8"))
                except Exception as e:
                    self.wfile.write(f"Error serving orderbook: {e}".encode("utf-8"))
                return

        portfolio = PortfolioManager.load_portfolio()
        
        # 1. Accounts Summary Endpoint
        if path == "/api/v1/accounts/summary":
            # Calculate total valuation dynamically based on current market prices
            total_equity = 0.0
            positions_summary = []
            
            for symbol, pos in portfolio["positions"].items():
                curr_price = get_current_price(symbol)
                mkt_val = pos["qty"] * curr_price
                total_equity += mkt_val
                
            total_value = portfolio["cash"] + total_equity
            
            summary = {
                "account_id": "MOCK-SCHWAB-12345",
                "cash_balance": round(portfolio["cash"], 2),
                "total_equity_value": round(total_equity, 2),
                "account_net_worth": round(total_value, 2),
                "day_change_percent": -0.08,
                "timestamp": datetime.utcnow().isoformat()
            }
            self._send_json(200, summary)
            
        # 2. Account Positions Endpoint
        elif path == "/api/v1/accounts/positions":
            positions = []
            for symbol, pos in portfolio["positions"].items():
                curr_price = get_current_price(symbol)
                mkt_val = pos["qty"] * curr_price
                unrealized_pl = mkt_val - pos["cost_basis"]
                
                positions.append({
                    "symbol": symbol,
                    "quantity": pos["qty"],
                    "current_price": curr_price,
                    "market_value": round(mkt_val, 2),
                    "cost_basis": pos["cost_basis"],
                    "unrealized_pl": round(unrealized_pl, 2),
                    "unrealized_pl_percent": round((unrealized_pl / pos["cost_basis"] * 100) if pos["cost_basis"] > 0 else 0, 2)
                })
            self._send_json(200, positions)
            
        # 3. Active Orders Endpoint
        elif path == "/api/v1/orders":
            self._send_json(200, portfolio["orders"])
            
        else:
            self._send_json(404, {"error": "Not Found"})

    def do_POST(self):
        parsed_url = urllib.parse.urlparse(self.path)
        path = parsed_url.path
        
        content_length = int(self.headers.get("Content-Length", 0))
        post_data = self.rfile.read(content_length).decode("utf-8")
        
        portfolio = PortfolioManager.load_portfolio()
        
        # 1. Place Order Endpoint
        if path == "/api/v1/orders/place":
            try:
                body = json.loads(post_data)
                symbol = body["symbol"].upper()
                qty = float(body["qty"])
                action = body["action"].upper() # BUY or SELL
                order_type = body.get("type", "MARKET").upper()
                limit_price = body.get("price")
                
                if action not in ["BUY", "SELL"]:
                    return self._send_json(400, {"error": "Invalid action. Must be BUY or SELL"})
                if qty <= 0:
                    return self._send_json(400, {"error": "Quantity must be positive"})
                    
                curr_price = get_current_price(symbol)
                exec_price = limit_price if (order_type == "LIMIT" and limit_price) else curr_price
                
                # Check Cash logic for BUY orders
                if action == "BUY":
                    cost = qty * exec_price
                    if cost > portfolio["cash"]:
                        return self._send_json(400, {"error": f"Insufficient funds. Order cost: ${cost:.2f}, Cash: ${portfolio['cash']:.2f}"})
                
                # Check Holdings logic for SELL orders
                if action == "SELL":
                    held_qty = portfolio["positions"].get(symbol, {}).get("qty", 0.0)
                    if held_qty < qty:
                        return self._send_json(400, {"error": f"Insufficient shares to sell {symbol}. Held: {held_qty}, Order: {qty}"})
                        
                order_id = f"ORD-{int(time.time() * 1000)}"
                
                new_order = {
                    "order_id": order_id,
                    "symbol": symbol,
                    "qty": qty,
                    "action": action,
                    "type": order_type,
                    "price": exec_price,
                    "status": "PENDING" if order_type == "LIMIT" else "FILLED",
                    "timestamp": datetime.utcnow().isoformat()
                }
                
                if new_order["status"] == "FILLED":
                    # Execute execution updates
                    if action == "BUY":
                        portfolio["cash"] -= (qty * exec_price)
                        pos = portfolio["positions"].get(symbol, {"qty": 0.0, "cost_basis": 0.0})
                        pos["qty"] += qty
                        pos["cost_basis"] += (qty * exec_price)
                        pos["price"] = exec_price
                        portfolio["positions"][symbol] = pos
                    else: # SELL
                        portfolio["cash"] += (qty * exec_price)
                        pos = portfolio["positions"][symbol]
                        # Reduce cost basis proportionally
                        avg_cost = pos["cost_basis"] / pos["qty"]
                        pos["qty"] -= qty
                        pos["cost_basis"] -= (qty * avg_cost)
                        if pos["qty"] <= 0:
                            del portfolio["positions"][symbol]
                            
                    portfolio["history"].append(new_order)
                    # Trigger Telegram Notification
                    telegram_notifier.send_trade_signal(symbol, qty, action, exec_price)
                else:
                    portfolio["orders"].append(new_order)
                    
                PortfolioManager.save_portfolio(portfolio)
                self._send_json(200, {
                    "status": "SUCCESS",
                    "message": f"Order {order_id} processed.",
                    "order": new_order
                })
            except Exception as e:
                self._send_json(400, {"error": f"Invalid payload: {e}"})
                
        # 2. Cancel Order Endpoint
        elif path == "/api/v1/orders/cancel":
            try:
                body = json.loads(post_data)
                order_id = body["order_id"]
                
                matched = [o for o in portfolio["orders"] if o["order_id"] == order_id]
                if not matched:
                    return self._send_json(404, {"error": "Pending order not found"})
                    
                portfolio["orders"] = [o for o in portfolio["orders"] if o["order_id"] != order_id]
                PortfolioManager.save_portfolio(portfolio)
                self._send_json(200, {"status": "SUCCESS", "message": f"Order {order_id} cancelled successfully."})
            except Exception as e:
                self._send_json(400, {"error": f"Invalid payload: {e}"})
        else:
            self._send_json(404, {"error": "Not Found"})

def run_server():
    server_address = ("", PORT)
    httpd = HTTPServer(server_address, SchwabRequestHandler)
    print(f"Schwab Mock Paper API server running on port {PORT}...")
    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down server...")
        httpd.server_close()

if __name__ == "__main__":
    run_server()
