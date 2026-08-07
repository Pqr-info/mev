from dataclasses import dataclass
from typing import Dict, Any

@dataclass
class PositionState:
    symbol: str
    side: str  # "long" / "short"
    qty: float
    entry_price: float
    current_price: float
    unrealized_pnl_pct: float
    holding_days: float

@dataclass
class SignalSummary:
    symbol: str
    liquidation_risk_score: float
    wave_continuation_score: float
    snipe_opportunity_score: float
    details: dict

class PositionAnalyzer:
    def __init__(self, signal_sources: Dict[str, Any]):
        self.signal_sources = signal_sources  # map of algo name to a signal generator

    def analyze_position(self, pos: PositionState, regime: Any, market_context: dict) -> SignalSummary:
        ml_signal = self.signal_sources.get("ml").get_signal(pos.symbol) if "ml" in self.signal_sources else 0.0
        momentum_signal = self.signal_sources.get("momentum").get_signal(pos.symbol) if "momentum" in self.signal_sources else 0.0
        mr_signal = self.signal_sources.get("mean_reversion").get_signal(pos.symbol) if "mean_reversion" in self.signal_sources else 0.0
        stat_arb_signal = self.signal_sources.get("stat_arb").get_signal(pos.symbol) if "stat_arb" in self.signal_sources else 0.0

        liquidation_risk = 0.0
        wave_continuation = 0.0
        snipe_opportunity = 0.0

        # Liquidation risk
        if pos.unrealized_pnl_pct < -market_context.get("max_dd_pct", 0.05):
            liquidation_risk += 0.4
        if getattr(regime, "trend", "") == "down" and pos.side == "long":
            liquidation_risk += 0.3
        if ml_signal < 0 and momentum_signal < 0:
            liquidation_risk += 0.3

        # Wave continuation
        if getattr(regime, "trend", "") == "up" and pos.side == "long":
            wave_continuation += 0.4
        if ml_signal > 0 and momentum_signal > 0:
            wave_continuation += 0.4
        if market_context.get("volatility_in_range", False):
            wave_continuation += 0.2

        # Snipe opportunity (for symbols you don't hold or small size)
        if pos.qty == 0 or pos.qty < market_context.get("min_size_for_full", 1.0):
            if ml_signal > market_context.get("ml_threshold", 0.5) and momentum_signal > market_context.get("mom_threshold", 0.5):
                snipe_opportunity += 0.5
            if mr_signal < -market_context.get("mr_z_threshold", 2.0):
                snipe_opportunity += 0.3
            if market_context.get("spread_bps", 100) < market_context.get("max_spread_bps", 5.0):
                snipe_opportunity += 0.2

        return SignalSummary(
            symbol=pos.symbol,
            liquidation_risk_score=liquidation_risk,
            wave_continuation_score=wave_continuation,
            snipe_opportunity_score=snipe_opportunity,
            details={
                "ml_signal": ml_signal,
                "momentum_signal": momentum_signal,
                "mr_signal": mr_signal,
                "stat_arb_signal": stat_arb_signal,
                "regime": regime.__dict__ if hasattr(regime, '__dict__') else regime,
                "position": pos.__dict__,
            }
        )
