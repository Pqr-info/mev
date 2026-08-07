from dataclasses import dataclass
from typing import List
from collections import defaultdict
from .position_analyzer import SignalSummary

@dataclass
class PositionExposure(SignalSummary):
    notional_value: float
    regime: str
    strategy: str

@dataclass
class CapitalSnapshot:
    exposures: List[PositionExposure]
    cash_balance: float
    total_equity: float

def build_capital_heatmap(snapshot: CapitalSnapshot) -> dict:
    by_regime = defaultdict(float)
    by_strategy = defaultdict(float)
    by_liquidation_risk = defaultdict(float)
    by_wave_state = defaultdict(float)
    by_snipe_state = defaultdict(float)

    for exp in snapshot.exposures:
        by_regime[exp.regime] += exp.notional_value
        by_strategy[exp.strategy] += exp.notional_value

        # Bucket liquidation risk
        if exp.liquidation_risk_score >= 0.7:
            by_liquidation_risk["high"] += exp.notional_value
        elif exp.liquidation_risk_score >= 0.4:
            by_liquidation_risk["medium"] += exp.notional_value
        else:
            by_liquidation_risk["low"] += exp.notional_value

        # Bucket wave continuation
        if exp.wave_continuation_score >= 0.7:
            by_wave_state["strong_wave"] += exp.notional_value
        elif exp.wave_continuation_score >= 0.4:
            by_wave_state["moderate_wave"] += exp.notional_value
        else:
            by_wave_state["weak_or_none"] += exp.notional_value

        # Bucket snipe opportunity (for symbols you *could* enter more)
        if exp.snipe_opportunity_score >= 0.7:
            by_snipe_state["high_opportunity"] += exp.notional_value
        elif exp.snipe_opportunity_score >= 0.4:
            by_snipe_state["medium_opportunity"] += exp.notional_value
        else:
            by_snipe_state["low_opportunity"] += exp.notional_value

    return {
        "by_regime": dict(by_regime),
        "by_strategy": dict(by_strategy),
        "by_liquidation_risk": dict(by_liquidation_risk),
        "by_wave_state": dict(by_wave_state),
        "by_snipe_state": dict(by_snipe_state),
        "cash_balance": snapshot.cash_balance,
        "total_equity": snapshot.total_equity,
    }
