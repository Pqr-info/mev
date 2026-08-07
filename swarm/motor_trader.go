package main

import (
	"fmt"
	"time"
)

type TradeMutation struct {
	Kind      string  `json:"kind"`
	Symbol    string  `json:"symbol,omitempty"`
	Qty       float64 `json:"qty,omitempty"`
	CostBasis float64 `json:"cost_basis,omitempty"`
	Amount    float64 `json:"amount,omitempty"`
}

type TradeVerdict struct {
	Approved  bool   `json:"approved"`
	Rationale string `json:"rationale"`
}

type TradeEntry struct {
	LineageID string        `json:"lineage_id"`
	Timestamp string        `json:"timestamp"`
	Mutation  TradeMutation `json:"mutation"`
	Verdict   TradeVerdict  `json:"verdict"`
}

type PortfolioSnapshot struct {
	Timestamp   time.Time         `json:"timestamp"`
	LineageID   string            `json:"lineage_id"`
	Portfolio   Portfolio         `json:"portfolio"`
}

type MotorTrader struct {
	PortfolioEngine *TemporalPortfolioEngine
	ConsensusEngine *ConsensusEngine
	Bridge          *TraderBridge
	MaxExposure     float64
	CashFloor       float64
}

func NewMotorTrader(pe *TemporalPortfolioEngine, ce *ConsensusEngine, bridge *TraderBridge) *MotorTrader {
	return &MotorTrader{
		PortfolioEngine: pe,
		ConsensusEngine: ce,
		Bridge:          bridge,
		MaxExposure:     5000.0, // Constitutional Symbol Exposure limit
		CashFloor:       1000.0, // Reserve floor
	}
}

// Tick evaluates sensory inputs and generates mutation proposals
func (mt *MotorTrader) Tick(accountID string, config ValidatorConfig) []TradeEntry {
	p := mt.PortfolioEngine.GetOrCreate(accountID)
	var generated []TradeEntry

	// Heuristic 1: Exposure Drift rebalancing (reduce positions exceeding MT limit)
	for symbol, pos := range p.Positions {
		exposure := pos.Qty * pos.CostBasis
		if exposure > mt.MaxExposure {
			targetQty := mt.MaxExposure / pos.CostBasis
			reduceQty := pos.Qty - targetQty

			if reduceQty > 0 {
				mut := FinancialMutation{
					Kind:      PortfolioRemove,
					Symbol:    symbol,
					Qty:       reduceQty,
					CostBasis: pos.CostBasis,
				}

				propID := fmt.Sprintf("motor-drift-%s-%d", symbol, time.Now().UnixNano())
				verdict := mt.ConsensusEngine.DecideFinancial(p, mut, config, nil)

				generated = append(generated, TradeEntry{
					LineageID: propID,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Mutation: TradeMutation{
						Kind:      string(mut.Kind),
						Symbol:    mut.Symbol,
						Qty:       mut.Qty,
						CostBasis: mut.CostBasis,
					},
					Verdict: TradeVerdict{
						Approved:  verdict.Approved,
						Rationale: fmt.Sprintf("[Motor Heuristic] Exposure drift rebalance: %s", verdict.Reason),
					},
				})
			}
		}
	}

	// Heuristic 2: Cash floor warning & rebalance proposal
	if p.Cash < mt.CashFloor {
		mut := FinancialMutation{
			Kind:   CashAdjust,
			Amount: mt.CashFloor - p.Cash,
		}

		propID := fmt.Sprintf("motor-floor-%d", time.Now().UnixNano())
		verdict := mt.ConsensusEngine.DecideFinancial(p, mut, config, nil)

		generated = append(generated, TradeEntry{
			LineageID: propID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Mutation: TradeMutation{
				Kind:   string(mut.Kind),
				Amount: mut.Amount,
			},
			Verdict: TradeVerdict{
				Approved:  verdict.Approved,
				Rationale: fmt.Sprintf("[Motor Heuristic] Cash floor violation: %s", verdict.Reason),
			},
		})
	}

	return generated
}

// TemporalTick analyzes historical snapshots to enforce self-correction feedback
func (mt *MotorTrader) TemporalTick(timeline []PortfolioSnapshot) []TradeEntry {
	var generated []TradeEntry
	if len(timeline) < 3 {
		return generated
	}

	// Detect leverage oscillation (leverage going up/down rapidly over last 3 snapshots)
	l1 := timeline[len(timeline)-1].Portfolio.Risk.CurrentLeverage
	l2 := timeline[len(timeline)-2].Portfolio.Risk.CurrentLeverage
	l3 := timeline[len(timeline)-3].Portfolio.Risk.CurrentLeverage

	if l1 > 2.0 && l2 < 1.0 && l3 > 2.0 {
		// Propose risk delta adjustment to stabilize
		mut := FinancialMutation{
			Kind:  RiskProfileUpdate,
			Delta: -0.5,
		}

		propID := fmt.Sprintf("motor-temporal-stabilize-%d", time.Now().UnixNano())
		generated = append(generated, TradeEntry{
			LineageID: propID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Mutation: TradeMutation{
				Kind:   string(mut.Kind),
				Amount: mut.Delta,
			},
			Verdict: TradeVerdict{
				Approved:  true,
				Rationale: "[Motor Heuristic] Leverage oscillation detected. Restricting leverage limit.",
			},
		})
	}

	return generated
}
