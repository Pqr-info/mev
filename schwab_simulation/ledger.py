import threading
import time
from collections import defaultdict
from dataclasses import dataclass
from typing import List, Dict

# Simple immutable ledger for simulated Schwab orders

@dataclass(frozen=True)
class Order:
    order_id: str
    symbol: str
    side: str  # 'buy' or 'sell'
    quantity: int
    price: float
    timestamp: float

class Ledger:
    def __init__(self, latency_seconds: float = 0.5):
        self._orders: List[Order] = []
        self._positions: Dict[str, int] = defaultdict(int)
        self.latency = latency_seconds
        self._lock = threading.Lock()

    def submit_order(self, order: Order) -> None:
        """Submit an order with simulated latency. The ledger is immutable –
        orders are appended to the list and never modified.
        """
        def process():
            time.sleep(self.latency)
            with self._lock:
                self._orders.append(order)
                qty = order.quantity if order.side == 'buy' else -order.quantity
                self._positions[order.symbol] += qty
        threading.Thread(target=process, daemon=True).start()

    @property
    def orders(self) -> List[Order]:
        # Return a copy to preserve immutability
        with self._lock:
            return list(self._orders)

    def get_position(self, symbol: str) -> int:
        with self._lock:
            return self._positions.get(symbol, 0)

    def all_positions(self) -> Dict[str, int]:
        with self._lock:
            return dict(self._positions)

# Singleton ledger instance used by the engine
ledger = Ledger(latency_seconds=0.5)
