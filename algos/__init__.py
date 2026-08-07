# algos package initializer

"""Algorithm registry for the paper‑trading platform.

Each algorithm module defines a class with a `generate_signals` method that
receives a `market_data` dictionary (symbol → pandas DataFrame) and returns a
list of order dictionaries:

    {
        "broker": "schwab" | "alpaca",
        "symbol": str,
        "side": "buy" | "sell",
        "quantity": int,
        "price": float
    }

The registry makes it easy for the engine to discover and run all available
strategies.
"""

from importlib import import_module
from typing import Dict, List, Callable

# Mapping from algorithm name to factory callable returning an instance
ALGORITHM_REGISTRY: Dict[str, Callable[[], object]] = {}

def register(name: str):
    """Decorator to register an algorithm class under `name`."""
    def decorator(cls):
        ALGORITHM_REGISTRY[name] = cls
        return cls
    return decorator

# Dynamically import all algorithm modules so they register themselves
for module_name in ["mean_reversion", "momentum", "stat_arbitrage", "ml_predictor"]:
    try:
        import_module(f"algos.{module_name}")
    except Exception:
        # In a production system we would log this; here we silently ignore.
        pass

def get_algorithms() -> List[object]:
    """Instantiate all registered algorithms and return the list."""
    return [cls() for cls in ALGORITHM_REGISTRY.values()]

