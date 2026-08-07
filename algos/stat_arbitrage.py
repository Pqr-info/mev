from . import register
import pandas as pd

@register('stat_arbitrage')
class StatisticalArbitrageStrategy:
    """Simple pair‑trading statistical arbitrage.
    Assumes two correlated symbols are passed in `market_data` with matching keys.
    If the price ratio deviates from its historic mean by `threshold`, we take
    opposite positions to bet on convergence.
    """

    def __init__(self, symbol_a: str = 'AAPL', symbol_b: str = 'MSFT', lookback: int = 30, threshold: float = 0.02, max_qty: int = 5):
        self.symbol_a = symbol_a
        self.symbol_b = symbol_b
        self.lookback = lookback
        self.threshold = threshold
        self.max_qty = max_qty

    def generate_signals(self, market_data: dict) -> list:
        # Ensure both symbols are present
        if self.symbol_a not in market_data or self.symbol_b not in market_data:
            return []
        df_a = market_data[self.symbol_a]
        df_b = market_data[self.symbol_b]
        if len(df_a) < self.lookback or len(df_b) < self.lookback:
            return []
        # Compute price ratio series
        ratio = df_a['close'].iloc[-self.lookback:] / df_b['close'].iloc[-self.lookback:]
        mean_ratio = ratio.mean()
        current_ratio = ratio.iloc[-1]
        orders = []
        # If ratio is higher than mean + threshold => sell A, buy B
        if current_ratio > (1 + self.threshold) * mean_ratio:
            price_a = df_a['close'].iloc[-1]
            price_b = df_b['close'].iloc[-1]
            orders.append({
                'broker': 'schwab',
                'symbol': self.symbol_a,
                'side': 'sell',
                'quantity': self.max_qty,
                'price': float(price_a)
            })
            orders.append({
                'broker': 'schwab',
                'symbol': self.symbol_b,
                'side': 'buy',
                'quantity': self.max_qty,
                'price': float(price_b)
            })
        # If ratio is lower than mean - threshold => buy A, sell B
        elif current_ratio < (1 - self.threshold) * mean_ratio:
            price_a = df_a['close'].iloc[-1]
            price_b = df_b['close'].iloc[-1]
            orders.append({
                'broker': 'schwab',
                'symbol': self.symbol_a,
                'side': 'buy',
                'quantity': self.max_qty,
                'price': float(price_a)
            })
            orders.append({
                'broker': 'schwab',
                'symbol': self.symbol_b,
                'side': 'sell',
                'quantity': self.max_qty,
                'price': float(price_b)
            })
        return orders
