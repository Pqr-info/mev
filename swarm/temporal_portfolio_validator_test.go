package main

import (
	"testing"
)

func TestFinancialValidatorConstraints(t *testing.T) {
	pe := NewTemporalPortfolioEngine()
	account := "test-account"
	p := pe.GetOrCreate(account)

	config := DefaultValidatorConfig()
	config.MaxExposurePerSymbol["OMCL"] = 5000.0

	// Initialize cash
	mutCash := FinancialMutation{
		Kind:   CashAdjust,
		Amount: 10000.0,
	}
	_ = pe.ApplyMutation(account, mutCash)

	// 1. Validate Forbidden Instruments
	mutForbidden := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "FORBIDDEN_COIN",
		Qty:       10.0,
		CostBasis: 10.0,
	}
	allowed, reason := ValidateFinancialMutation(p, mutForbidden, config)
	if allowed {
		t.Errorf("expected forbidden instrument validation to block trade, reason: %s", reason)
	}

	// 2. Validate Cash Floor Breach
	mutCashBreach := FinancialMutation{
		Kind:   CashAdjust,
		Amount: -12000.0,
	}
	allowed, reason = ValidateFinancialMutation(p, mutCashBreach, config)
	if allowed {
		t.Errorf("expected cash floor breach to block, reason: %s", reason)
	}

	// 3. Validate Symbol Exposure Breach
	mutExposureBreach := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "OMCL",
		Qty:       100.0,
		CostBasis: 60.0, // 6000.0 exposure exceeds 5000.0
	}
	allowed, reason = ValidateFinancialMutation(p, mutExposureBreach, config)
	if allowed {
		t.Errorf("expected symbol exposure breach to block, reason: %s", reason)
	}

	// 4. Validate Valid Trade Allowed
	mutValid := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "OMCL",
		Qty:       50.0,
		CostBasis: 50.0, // 2500.0 exposure is valid
	}
	allowed, reason = ValidateFinancialMutation(p, mutValid, config)
	if !allowed {
		t.Errorf("expected valid trade to pass, blocked with: %s", reason)
	}

	// 5. Validate Risk Delta Breach
	mutRiskBreach := FinancialMutation{
		Kind:  RiskProfileUpdate,
		Delta: 0.9, // Projected risk 0.9 exceeds max 0.8
	}
	allowed, reason = ValidateFinancialMutation(p, mutRiskBreach, config)
	if allowed {
		t.Errorf("expected risk score breach to block, reason: %s", reason)
	}
}
