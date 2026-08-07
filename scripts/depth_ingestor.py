from dataclasses import dataclass
from datetime import datetime

@dataclass
class DepthLevel:
    price: float
    size: float
    side: str   # "bid" or "ask"

@dataclass
class SourceDepthSnapshot:
    symbol: str
    source: str          # "schwab", "alpaca", "source_c"
    timestamp: datetime
    bids: list[DepthLevel]  # sorted best→worse
    asks: list[DepthLevel]  # sorted best→worse

@dataclass
class AggregatedDepth:
    symbol: str
    timestamp: datetime
    sources: list[SourceDepthSnapshot]
    consolidated_best_bid: float
    consolidated_best_ask: float
    consolidated_spread: float
    per_source_spread: dict   # {source: spread}
    per_source_liquidity: dict # {source: total_depth_size}
    cross_source_widespread: dict # {("schwab","alpaca"): delta, ...}

class DepthIngestor:
    def __init__(self, source_name, client):
        self.source_name = source_name
        self.client = client

    def stream_depth(self, symbols):
        # yields SourceDepthSnapshot in real time
        pass

def aggregate_depth(symbol: str, snapshots: list[SourceDepthSnapshot]) -> AggregatedDepth:
    ts = max(s.timestamp for s in snapshots)

    per_source_spread = {}
    per_source_liquidity = {}
    for s in snapshots:
        best_bid = s.bids[0].price if s.bids else None
        best_ask = s.asks[0].price if s.asks else None
        if best_bid is not None and best_ask is not None:
            per_source_spread[s.source] = best_ask - best_bid
        total_depth = sum(l.size for l in s.bids + s.asks)
        per_source_liquidity[s.source] = total_depth

    # consolidated best bid/ask
    all_bids = [s.bids[0].price for s in snapshots if s.bids]
    all_asks = [s.asks[0].price for s in snapshots if s.asks]
    consolidated_best_bid = max(all_bids) if all_bids else None
    consolidated_best_ask = min(all_asks) if all_asks else None
    consolidated_spread = (
        consolidated_best_ask - consolidated_best_bid
        if consolidated_best_bid is not None and consolidated_best_ask is not None
        else None
    )

    # cross-source widespread
    cross_source_widespread = {}
    for a in snapshots:
        for b in snapshots:
            if a.source == b.source:
                continue
            if a.bids and b.asks:
                cross_source_widespread[f"{a.source}->{b.source}"] = b.asks[0].price - a.bids[0].price

    return AggregatedDepth(
        symbol=symbol,
        timestamp=ts,
        sources=snapshots,
        consolidated_best_bid=consolidated_best_bid,
        consolidated_best_ask=consolidated_best_ask,
        consolidated_spread=consolidated_spread,
        per_source_spread=per_source_spread,
        per_source_liquidity=per_source_liquidity,
        cross_source_widespread=cross_source_widespread,
    )
