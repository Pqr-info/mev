package main

import (
	"testing"
)

func TestPortfolioApplyMutation(t *testing.T) {
	pe := NewTemporalPortfolioEngine()
	account := "test-account"
	p := pe.GetOrCreate(account)

	// Test CashAdjust
	mutCash := FinancialMutation{
		Kind:   CashAdjust,
		Amount: 10000.0,
	}
	if err := pe.ApplyMutation(account, mutCash); err != nil {
		t.Fatalf("failed to adjust cash: %v", err)
	}
	if p.Cash != 10000.0 {
		t.Errorf("expected cash 10000, got %f", p.Cash)
	}

	// Test PortfolioAdd
	mutAdd := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "OMCL",
		Qty:       10.0,
		CostBasis: 50.0,
	}
	if err := pe.ApplyMutation(account, mutAdd); err != nil {
		t.Fatalf("failed to add asset: %v", err)
	}
	pos := p.Positions["OMCL"]
	if pos.Qty != 10.0 || pos.CostBasis != 50.0 {
		t.Errorf("expected 10 OMCL at 50, got qty=%f, cost=%f", pos.Qty, pos.CostBasis)
	}

	// Test PortfolioAdd cumulative cost basis
	mutAddMore := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "OMCL",
		Qty:       10.0,
		CostBasis: 60.0,
	}
	if err := pe.ApplyMutation(account, mutAddMore); err != nil {
		t.Fatalf("failed to add more asset: %v", err)
	}
	pos = p.Positions["OMCL"]
	if pos.Qty != 20.0 || pos.CostBasis != 55.0 {
		t.Errorf("expected 20 OMCL at average cost 55, got qty=%f, cost=%f", pos.Qty, pos.CostBasis)
	}

	// Test PortfolioRemove
	mutRemove := FinancialMutation{
		Kind:   PortfolioRemove,
		Symbol: "OMCL",
		Qty:    5.0,
	}
	if err := pe.ApplyMutation(account, mutRemove); err != nil {
		t.Fatalf("failed to remove asset: %v", err)
	}
	pos = p.Positions["OMCL"]
	if pos.Qty != 15.0 {
		t.Errorf("expected 15 OMCL remaining, got %f", pos.Qty)
	}

	// Test RiskProfileUpdate
	mutRisk := FinancialMutation{
		Kind:  RiskProfileUpdate,
		Delta: 0.15,
	}
	if err := pe.ApplyMutation(account, mutRisk); err != nil {
		t.Fatalf("failed to update risk: %v", err)
	}
	if p.Risk.CurrentRiskScore != 0.15 {
		t.Errorf("expected risk score 0.15, got %f", p.Risk.CurrentRiskScore)
	}
}
