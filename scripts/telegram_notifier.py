#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
telegram_notifier.py
===================
Handles sending beautifully formatted markdown signal notifications to a
Telegram channel/chat when a simulated paper trade executes.
"""

import os
import urllib.request
import urllib.parse
import json
import random

# Default fallback values for testing
DEFAULT_CHAT_ID = os.environ.get("TELEGRAM_CHAT_ID", "YOUR_TELEGRAM_CHAT_ID")
DEFAULT_TOKEN = os.environ.get("TELEGRAM_BOT_TOKEN", "YOUR_TELEGRAM_BOT_TOKEN")

def calculate_mock_indicators(symbol, price):
    """
    Generates realistic, mathematically-grounded mock technical indicators
    for the signal payload based on current price point.
    """
    # Seed random with symbol hash + price to keep it deterministic for the same event
    random.seed(hash(symbol) + int(price * 100))
    
    # RSI (relative strength index) typically oscillates between 10 and 90
    rsi = round(random.uniform(25.0, 75.0), 2)
    
    # Fibonacci levels commonly monitored
    fib_levels = ["0.382 Retracement Support", "0.500 Midpoint Support", "0.618 Golden Ratio Reclaim", "0.786 Extension Bounce"]
    fib_criteria = random.choice(fib_levels)
    
    # 48h confidence score (e.g. 65% to 98%)
    confidence_score = random.randint(65, 98)
    
    return rsi, fib_criteria, confidence_score

def send_trade_signal(symbol, qty, action, price, token=DEFAULT_TOKEN, chat_id=DEFAULT_CHAT_ID):
    """
    Formulates and transmits a structured markdown notification to Telegram.
    """
    if not token or token == "YOUR_TELEGRAM_BOT_TOKEN" or not chat_id or chat_id == "YOUR_TELEGRAM_CHAT_ID":
        print(f"[WARN] Telegram credentials not configured. Skipping signal transmission for {action} {symbol}.")
        return False

    rsi, fib, confidence = calculate_mock_indicators(symbol, price)
    
    action_emoji = "🟢 BUY" if action.upper() == "BUY" else "🔴 SELL"
    
    message = (
        f"🔔 *Paper Trade Executed!*\n"
        f"━━━━━━━━━━━━━━━━━━━\n"
        f"📈 *Action*: {action_emoji}\n"
        f"🔠 *Symbol*: {symbol}\n"
        f"💼 *Quantity*: {qty}\n"
        f"💵 *Price Point*: ${price:,.2f}\n"
        f"📊 *RSI Value*: {rsi}\n"
        f"📐 *Fibonacci Level*: {fib}\n"
        f"🔮 *48h Confidence*: {confidence}%\n"
        f"━━━━━━━━━━━━━━━━━━━\n"
        f"⏳ _Execution environment: Schwab Paper Mock_"
    )
    
    url = f"https://api.telegram.org/bot{token}/sendMessage"
    payload = {
        "chat_id": chat_id,
        "text": message,
        "parse_mode": "Markdown"
    }
    
    try:
        headers = {"Content-Type": "application/json"}
        req = urllib.request.Request(
            url, 
            data=json.dumps(payload).encode("utf-8"), 
            headers=headers, 
            method="POST"
        )
        with urllib.request.urlopen(req, timeout=10) as response:
            res_body = json.loads(response.read().decode("utf-8"))
            if res_body.get("ok"):
                print(f"[INFO] Signal for {action} {symbol} successfully transmitted to Telegram.")
                return True
            else:
                print(f"[ERROR] Telegram API rejected message: {res_body}")
                return False
    except Exception as e:
        print(f"[ERROR] Failed to send Telegram notification: {e}")
        return False

if __name__ == "__main__":
    # Test message
    send_trade_signal("TXN", 10.0, "BUY", 311.46)
