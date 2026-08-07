import json
import os
from datetime import datetime, timezone

def format_timestamp_for_filename(ts: datetime) -> str:
    return ts.astimezone(timezone.utc).isoformat().replace(":", "-")

def microstructure_path(root: str, ts: datetime) -> str:
    ts = ts.astimezone(timezone.utc)
    return os.path.join(
        root,
        "microstructure",
        f"{ts.year:04d}",
        f"{ts.month:02d}",
        f"{ts.day:02d}",
        f"{ts.hour:02d}",
        f"{ts.minute:02d}",
        f"{format_timestamp_for_filename(ts)}.json",
    )

def write_microstructure_event(root: str, ts: datetime, event: dict, recommended_trades: list[dict] = None) -> None:
    path = microstructure_path(root, ts)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    
    if recommended_trades is not None:
        event["recommended_trades"] = recommended_trades
        
    with open(path, "w", encoding="utf-8") as f:
        json.dump(event, f, indent=2)
