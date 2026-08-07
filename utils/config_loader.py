import yaml
import os
from pathlib import Path

DEFAULT_CONFIG = {
    "assets": ["AAPL", "MSFT", "GOOGL", "SPY"],
    "risk_limits": {
        "max_position_per_symbol": 100,
        "max_daily_loss": 1000
    },
    "schwab": {
        "latency_seconds": 0.5
    },
    "alpaca": {
        "base_url": "https://paper-api.alpaca.markets",
        "data_url": "https://data.alpaca.markets/v2"
    }
}

def load_config(config_path: str = "config.yaml") -> dict:
    """Load configuration from a YAML file, merging with defaults.
    Returns a dictionary with all needed settings.
    """
    config_file = Path(config_path)
    if not config_file.is_file():
        # No config file, return defaults
        return DEFAULT_CONFIG
    with open(config_file, "r", encoding="utf-8") as f:
        loaded = yaml.safe_load(f) or {}
    # Merge defaults with loaded (loaded overrides defaults)
    def merge(a, b):
        for k, v in b.items():
            if isinstance(v, dict) and k in a:
                merge(a[k], v)
            else:
                a[k] = v
    merged = DEFAULT_CONFIG.copy()
    merge(merged, loaded)
    return merged

