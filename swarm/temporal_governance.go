package main

import (
	"fmt"
	"math"
	"time"
)

type TemporalGovernanceDirective struct {
	Reason             string
	ReplayAllowed      bool
	TimeslipAllowed    bool
	MutationRateCap    float64
	FederationSyncFreq time.Duration
}

type TemporalConstraintExposure struct {
	MaxDriftPct float64
	WindowSize  int
}

type TemporalConstraintLeverage struct {
	MaxOscillation float64
	WindowSize     int
}

type TemporalConstraintRisk struct {
	MaxRiskDelta float64
	WindowSize   int
}

type TemporalConstraintCash struct {
	MaxBelowFloor int
}

type TemporalConstraints struct {
	Exposure TemporalConstraintExposure
	Leverage TemporalConstraintLeverage
	Risk     TemporalConstraintRisk
	Cash     TemporalConstraintCash
}

type TemporalGovernance struct {
	MaxExposure float64
	CashFloor   float64
	Constraints TemporalConstraints
}

func NewTemporalGovernance() *TemporalGovernance {
	return &TemporalGovernance{
		MaxExposure: 5000.0,
		CashFloor:   1000.0,
		Constraints: TemporalConstraints{
			Exposure: TemporalConstraintExposure{
				MaxDriftPct: 0.15, // 15% Max drift allowed
				WindowSize:  5,
			},
			Leverage: TemporalConstraintLeverage{
				MaxOscillation: 0.5,
				WindowSize:     5,
			},
			Risk: TemporalConstraintRisk{
				MaxRiskDelta: 0.1,
				WindowSize:   5,
			},
			Cash: TemporalConstraintCash{
				MaxBelowFloor: 3,
			},
		},
	}
}

func getSymbolExposure(p Portfolio) map[string]float64 {
	exposure := make(map[string]float64)
	for symbol, pos := range p.Positions {
		exposure[symbol] = pos.Qty * pos.CostBasis
	}
	return exposure
}

func getLeverage(p Portfolio) float64 {
	totalEquity := 0.0
	for _, pos := range p.Positions {
		totalEquity += pos.Qty * pos.CostBasis
	}
	netLiq := p.Cash + totalEquity
	if netLiq > 0 {
		return totalEquity / netLiq
	}
	return 0.0
}

func (tg *TemporalGovernance) Tick(timeline []PortfolioSnapshot) []TradeEntry {
	var generated []TradeEntry

	// Check Constraints
	generated = append(generated, tg.checkExposureDrift(timeline)...)
	generated = append(generated, tg.checkLeverageOscillation(timeline)...)
	generated = append(generated, tg.checkCashFloor(timeline)...)

	return generated
}

func (tg *TemporalGovernance) checkExposureDrift(timeline []PortfolioSnapshot) []TradeEntry {
	var generated []TradeEntry
	wSize := tg.Constraints.Exposure.WindowSize
	if len(timeline) < wSize {
		return generated
	}

	// Slice of the last WindowSize snapshots
	recent := timeline[len(timeline)-wSize:]
	first := getSymbolExposure(recent[0].Portfolio)
	last := getSymbolExposure(recent[len(recent)-1].Portfolio)

	for symbol, lastExp := range last {
		firstExp, exists := first[symbol]
		if !exists || firstExp == 0 {
			continue
		}

		drift := (lastExp - firstExp) / firstExp
		if drift > tg.Constraints.Exposure.MaxDriftPct {
			propID := fmt.Sprintf("gov-drift-%s-%d", symbol, time.Now().UnixNano())
			
			lastSnap := recent[len(recent)-1]
			var currentQty float64
			var costBasis float64
			for _, pos := range lastSnap.Portfolio.Positions {
				if pos.Symbol == symbol {
					currentQty = pos.Qty
					costBasis = pos.CostBasis
					break
				}
			}

			if currentQty > 0 {
				generated = append(generated, TradeEntry{
					LineageID: propID,
					Timestamp: time.Now().UTC().Format(time.RFC3339),
					Mutation: TradeMutation{
						Kind:      "portfolio_adjust",
						Symbol:    symbol,
						Qty:       currentQty * 0.9, // Reduce position by 10%
						CostBasis: costBasis,
					},
					Verdict: TradeVerdict{
						Approved:  true,
						Rationale: fmt.Sprintf("[Temporal Governance] Exposure drift of %s exceeded limit (Drift: %.2f%%). Reducing exposure.", symbol, drift*100),
					},
				})
			}
		}
	}

	return generated
}

func (tg *TemporalGovernance) checkLeverageOscillation(timeline []PortfolioSnapshot) []TradeEntry {
	var generated []TradeEntry
	wSize := tg.Constraints.Leverage.WindowSize
	if len(timeline) < wSize {
		return generated
	}

	recent := timeline[len(timeline)-wSize:]
	minL, maxL := math.MaxFloat64, -math.MaxFloat64

	for _, snap := range recent {
		L := getLeverage(snap.Portfolio)
		if L < minL {
			minL = L
		}
		if L > maxL {
			maxL = L
		}
	}

	if (maxL - minL) > tg.Constraints.Leverage.MaxOscillation {
		propID := fmt.Sprintf("gov-leverage-stabilize-%d", time.Now().UnixNano())
		generated = append(generated, TradeEntry{
			LineageID: propID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Mutation: TradeMutation{
				Kind:   "risk_profile_update",
				Amount: -0.05,
			},
			Verdict: TradeVerdict{
				Approved:  true,
				Rationale: fmt.Sprintf("[Temporal Governance] Leverage oscillation of %.2f exceeded limit of %.2f. Enforcing risk delta cap.", (maxL - minL), tg.Constraints.Leverage.MaxOscillation),
			},
		})
	}

	return generated
}

func (tg *TemporalGovernance) checkCashFloor(timeline []PortfolioSnapshot) []TradeEntry {
	var generated []TradeEntry
	wSize := tg.Constraints.Cash.MaxBelowFloor
	if len(timeline) < wSize {
		return generated
	}

	recent := timeline[len(timeline)-wSize:]
	belowCount := 0
	for _, snap := range recent {
		if snap.Portfolio.Cash < tg.CashFloor {
			belowCount++
		}
	}

	if belowCount >= wSize {
		propID := fmt.Sprintf("gov-cash-restructure-%d", time.Now().UnixNano())
		lastSnap := recent[len(recent)-1]
		generated = append(generated, TradeEntry{
			LineageID: propID,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Mutation: TradeMutation{
				Kind:   "cash_adjust",
				Amount: tg.CashFloor - lastSnap.Portfolio.Cash,
			},
			Verdict: TradeVerdict{
				Approved:  true,
				Rationale: fmt.Sprintf("[Temporal Governance] Cash persistently below floor for %d ticks. Triggering auto-funding / rebalance proposal.", belowCount),
			},
		})
	}

	return generated
}
