package main

import (
	"fmt"
)

// ValidatorConfig represents the constitutional constraints defined under Article II
type ValidatorConfig struct {
	CashFloor             float64            `json:"cash_floor"`
	MaxLeverage           float64            `json:"max_leverage"`
	MaxExposurePerSymbol  map[string]float64 `json:"max_exposure_per_symbol"`
	ForbiddenInstruments  []string           `json:"forbidden_instruments"`
	MaxAllowedRiskScore   float64            `json:"max_allowed_risk_score"`
}

// DefaultValidatorConfig returns the standard constitutional risk bounds
func DefaultValidatorConfig() ValidatorConfig {
	return ValidatorConfig{
		CashFloor:            0.0,
		MaxLeverage:          3.0,
		MaxExposurePerSymbol: make(map[string]float64),
		ForbiddenInstruments: []string{"FORBIDDEN_COIN", "SCAM_TOKEN"},
		MaxAllowedRiskScore:  0.8,
	}
}

// ValidateFinancialMutation checks a mutation against Article II constitutional constraints before consensus
func ValidateFinancialMutation(p *Portfolio, mut FinancialMutation, config ValidatorConfig) (bool, string) {
	// 1. Check for Forbidden Instruments
	if mut.Symbol != "" {
		for _, forbidden := range config.ForbiddenInstruments {
			if mut.Symbol == forbidden {
				return false, fmt.Sprintf("Article II violation: symbol %s is a forbidden instrument", mut.Symbol)
			}
		}
	}

	// 2. Check Cash Floor Constraint
	if mut.Kind == CashAdjust {
		projectedCash := p.Cash + mut.Amount
		if projectedCash < config.CashFloor {
			return false, fmt.Sprintf("Article II violation: cash balance %f would fall below the floor of %f", projectedCash, config.CashFloor)
		}
	}

	// 3. Check Exposure Bounds
	if mut.Kind == PortfolioAdd {
		// Calculate projected value for the symbol
		pos := p.Positions[mut.Symbol]
		projectedQty := pos.Qty + mut.Qty
		projectedCost := projectedQty * mut.CostBasis

		// If a limit is configured for this symbol
		if limit, exists := config.MaxExposurePerSymbol[mut.Symbol]; exists {
			if projectedCost > limit {
				return false, fmt.Sprintf("Article II violation: exposure for symbol %s (%f) would exceed limit %f", mut.Symbol, projectedCost, limit)
			}
		}

		// Calculate total portfolio value (cash + positions equity)
		totalEquity := 0.0
		for _, position := range p.Positions {
			totalEquity += position.Qty * position.CostBasis
		}
		totalEquity += mut.Qty * mut.CostBasis

		totalValue := p.Cash + totalEquity
		if totalValue > 0 {
			// Leverage check: total positions equity / total value
			projectedLeverage := totalEquity / totalValue
			if projectedLeverage > config.MaxLeverage {
				return false, fmt.Sprintf("Article II violation: projected leverage %f would exceed max limit %f", projectedLeverage, config.MaxLeverage)
			}
		}
	}

	// 4. Check Risk Delta Bounds
	if mut.Kind == RiskProfileUpdate {
		projectedRisk := p.Risk.CurrentRiskScore + mut.Delta
		if projectedRisk > config.MaxAllowedRiskScore {
			return false, fmt.Sprintf("Article II violation: projected risk score %f would exceed limits of %f", projectedRisk, config.MaxAllowedRiskScore)
		}
	}

	return true, "Allowed"
}
