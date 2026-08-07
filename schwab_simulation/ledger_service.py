import multiprocessing as mp
import time
from dataclasses import dataclass
from typing import Dict, Any
import json
import random
import os
from datetime import datetime

def sample_distribution(dist, rng=random):
    t = dist["type"]
    if t == "fixed":
        return dist["ms"]
    if t == "uniform":
        return rng.uniform(dist["min_ms"], dist["max_ms"])
    if t == "normal":
        val = rng.gauss(dist["mean_ms"], dist["std_ms"])
        if "min_ms" in dist:
            val = max(val, dist["min_ms"])
        if "max_ms" in dist:
            val = min(val, dist["max_ms"])
        return val
    if t == "exponential":
        val = rng.expovariate(dist["lambda"])
        if "min_ms" in dist:
            val = max(val, dist["min_ms"])
        if "max_ms" in dist:
            val = min(val, dist["max_ms"])
        return val
    raise ValueError(f"Unknown distribution type: {t}")

def compute_latency_ms(symbol, event_type, profiles, context) -> float:
    applicable_rules = []
    # Using datetime.utcnow() to simulate order timestamp since OrderEvent just has a float timestamp
    now = datetime.utcnow()
    for profile in profiles.get("profiles", []):
        for rule in profile.get("rules", []):
            scope = rule.get("scope", {})
            symbols = scope.get("symbols", ["*"])
            event_types = scope.get("event_types", [])
            
            symbol_match = ("*" in symbols) or (symbol in symbols)
            event_match = (not event_types) or (event_type in event_types)
            
            if symbol_match and event_match:
                applicable_rules.append(rule)
                
    if not applicable_rules:
        return 0.0
        
    rule = applicable_rules[-1]
    base = sample_distribution(rule["base_distribution"])
    total_delta = 0.0
    
    for mod in rule.get("modifiers", []):
        if mod["type"] == "by_load":
            load = context.get("current_load", 0)
            for th in mod.get("thresholds", []):
                if load <= th.get("max", float('inf')):
                    total_delta += th.get("delta_ms", 0)
                    break
    
    return base + total_delta

@dataclass(frozen=True)
class OrderEvent:
    order_id: str
    symbol: str
    side: str  # 'buy' or 'sell'
    quantity: int
    price: float
    timestamp: float

class LedgerService(mp.Process):
    """
    The single source of truth for Schwab simulation. Provides a serialized 
    intake lane to prevent race conditions.
    """
    def __init__(self, order_queue: mp.Queue, state_dict: Dict[str, Any]):
        super().__init__()
        self.order_queue = order_queue
        self.state_dict = state_dict
        self.daemon = True
        
        # Load latency profiles
        self.latency_profiles = {}
        config_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), 'config', 'latency_profiles.json')
        try:
            with open(config_path, 'r') as f:
                self.latency_profiles = json.load(f)
        except Exception as e:
            print(f"[Ledger Service] Could not load latency profiles: {e}")

    def run(self):
        print("[Ledger Service] Started.")
        # Local state to avoid locking on every read, syncs to state_dict periodically
        positions = {}
        orders = []

        while True:
            try:
                event = self.order_queue.get()
                if event is None:
                    # Poison pill to shutdown
                    break
                
                # Apply simulated stochastic latency
                context = {"current_load": 0} # simplified for now
                latency_ms = compute_latency_ms(event.symbol, "ORDER_SUBMITTED", self.latency_profiles, context)
                time.sleep(latency_ms / 1000.0)

                orders.append(event)
                
                qty = event.quantity if event.side == 'buy' else -event.quantity
                positions[event.symbol] = positions.get(event.symbol, 0) + qty
                
                # Update shared state dict for the dashboard/engine
                self.state_dict['positions'] = dict(positions)
                
                # Update latency telemetry
                self.state_dict['last_order_latency_ms'] = latency_ms
                
            except Exception as e:
                print(f"[Ledger Service] Error processing order: {e}")

        print("[Ledger Service] Shutting down.")
