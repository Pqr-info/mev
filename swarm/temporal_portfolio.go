package main

import (
	"errors"
	"fmt"
	"sync"
)

// Position represents a single asset holding in the portfolio
type Position struct {
	Symbol    string  `json:"symbol"`
	Qty       float64 `json:"qty"`
	CostBasis float64 `json:"cost_basis"`
}

// RiskProfile represents the risk limits and scores of the portfolio
type RiskProfile struct {
	MaxLeverage          float64            `json:"max_leverage"`
	CurrentLeverage      float64            `json:"current_leverage"`
	MaxExposurePerSymbol map[string]float64 `json:"max_exposure_per_symbol"`
	CurrentRiskScore     float64            `json:"current_risk_score"`
}

// Portfolio is the first-class ledger citizen for substrate.portfolio state
type Portfolio struct {
	AccountID string              `json:"account_id"`
	Positions map[string]Position `json:"positions"`
	Cash      float64             `json:"cash"`
	Risk      RiskProfile         `json:"risk"`
}

// TemporalPortfolioEngine governs the substrate.portfolio state transitions
type TemporalPortfolioEngine struct {
	mu         sync.RWMutex
	portfolios map[string]*Portfolio
}

func NewTemporalPortfolioEngine() *TemporalPortfolioEngine {
	return &TemporalPortfolioEngine{
		portfolios: make(map[string]*Portfolio),
	}
}

// GetOrCreate returns the portfolio for the given account or creates a default one
func (pe *TemporalPortfolioEngine) GetOrCreate(accountID string) *Portfolio {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	p, exists := pe.portfolios[accountID]
	if !exists {
		p = &Portfolio{
			AccountID: accountID,
			Positions: make(map[string]Position),
			Cash:      0.0,
			Risk: RiskProfile{
				MaxLeverage:          3.0,
				CurrentLeverage:      0.0,
				MaxExposurePerSymbol: make(map[string]float64),
				CurrentRiskScore:     0.0,
			},
		}
		pe.portfolios[accountID] = p
	}
	return p
}

// ApplyMutation performs state transitions from FinancialMutation to the canonical portfolio
func (pe *TemporalPortfolioEngine) ApplyMutation(accountID string, mut FinancialMutation) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	p, exists := pe.portfolios[accountID]
	if !exists {
		return fmt.Errorf("portfolio account %s not initialized", accountID)
	}

	switch mut.Kind {
	case PortfolioAdd:
		if mut.Symbol == "" {
			return errors.New("missing symbol for portfolio_add")
		}
		pos, hasPos := p.Positions[mut.Symbol]
		if hasPos {
			// Recalculate cost basis
			totalCost := (pos.Qty * pos.CostBasis) + (mut.Qty * mut.CostBasis)
			pos.Qty += mut.Qty
			if pos.Qty > 0 {
				pos.CostBasis = totalCost / pos.Qty
			}
			p.Positions[mut.Symbol] = pos
		} else {
			p.Positions[mut.Symbol] = Position{
				Symbol:    mut.Symbol,
				Qty:       mut.Qty,
				CostBasis: mut.CostBasis,
			}
		}

	case PortfolioRemove:
		if mut.Symbol == "" {
			return errors.New("missing symbol for portfolio_remove")
		}
		pos, hasPos := p.Positions[mut.Symbol]
		if !hasPos {
			return fmt.Errorf("symbol %s not found in portfolio for removal", mut.Symbol)
		}
		if pos.Qty < mut.Qty {
			return fmt.Errorf("insufficient quantity of %s to remove (held: %f, removing: %f)", mut.Symbol, pos.Qty, mut.Qty)
		}
		pos.Qty -= mut.Qty
		if pos.Qty <= 0 {
			delete(p.Positions, mut.Symbol)
		} else {
			p.Positions[mut.Symbol] = pos
		}

	case PortfolioAdjust:
		if mut.Symbol == "" {
			return errors.New("missing symbol for portfolio_adjust")
		}
		if mut.Qty <= 0 {
			delete(p.Positions, mut.Symbol)
		} else {
			p.Positions[mut.Symbol] = Position{
				Symbol:    mut.Symbol,
				Qty:       mut.Qty,
				CostBasis: mut.CostBasis,
			}
		}

	case CashAdjust:
		p.Cash += mut.Amount

	case RiskProfileUpdate:
		p.Risk.CurrentRiskScore += mut.Delta

	default:
		return fmt.Errorf("unrecognized financial mutation kind: %s", mut.Kind)
	}

	return nil
}
