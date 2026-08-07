package main

import (
	"testing"
)

func TestTraderBridgeIntegration(t *testing.T) {
	pe := NewTemporalPortfolioEngine()
	ce := NewConsensusEngine(2) // Quorum 2
	bridge := NewTraderBridge(pe, ce)

	account := "test-account"
	config := DefaultValidatorConfig()

	// 1. Initial Cash Injection (Approved by default)
	mutCash := FinancialMutation{
		Kind:   CashAdjust,
		Amount: 5000.0,
	}
	approved, reason, lineageID := bridge.ProposeTrade(account, mutCash, nil, config)
	if !approved {
		t.Fatalf("expected cash injection to be approved, got reason: %s", reason)
	}
	if lineageID == "" {
		t.Fatal("expected a valid lineage ID for approved trade")
	}

	p := pe.GetOrCreate(account)
	if p.Cash != 5000.0 {
		t.Errorf("expected portfolio cash 5000, got %f", p.Cash)
	}

	// Verify lineage block exists
	block, exists := bridge.GetLineage(lineageID)
	if !exists {
		t.Fatalf("lineage block %s not found in Biographer store", lineageID)
	}
	if block.Decision != "Approved" || block.Organ != "TRADER" {
		t.Errorf("invalid lineage block content: %+v", block)
	}

	// 2. Propose Invalid Trade (Forbidden Instrument) -> Rejected immediately
	mutForbidden := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "FORBIDDEN_COIN",
		Qty:       10.0,
		CostBasis: 10.0,
	}
	approved, reason, lineageID = bridge.ProposeTrade(account, mutForbidden, nil, config)
	if approved {
		t.Fatal("expected forbidden coin trade to be rejected")
	}
	if lineageID == "" {
		t.Fatal("expected a lineage ID even for rejected trade")
	}

	// Verify rejection lineage block exists
	rejectBlock, exists := bridge.GetLineage(lineageID)
	if !exists {
		t.Fatalf("rejection lineage block %s not found", lineageID)
	}
	if rejectBlock.Decision != "Rejected" {
		t.Errorf("expected lineage block to state 'Rejected', got '%s'", rejectBlock.Decision)
	}
}

func TestTraderBridgeExecute(t *testing.T) {
	pe := NewTemporalPortfolioEngine()
	ce := NewConsensusEngine(1)
	bridge := NewTraderBridge(pe, ce)
	config := DefaultValidatorConfig()

	intent := TraderIntent{
		AccountID: "test-acc-2",
		Symbol:    "OMCL",
		Qty:       10.0,
		CostBasis: 45.0,
		Kind:      PortfolioAdd,
	}

	result, err := bridge.Execute(intent, nil, config)
	if err != nil {
		t.Fatalf("bridge execute failed: %v", err)
	}

	if !result.Approved {
		t.Fatalf("expected intent approval, got rejected with rationale: %s", result.Rationale)
	}

	if result.LineageID == "" {
		t.Fatal("expected a valid lineage ID from bridge execution result")
	}

	p := pe.GetOrCreate("test-acc-2")
	pos, exists := p.Positions["OMCL"]
	if !exists || pos.Qty != 10.0 {
		t.Errorf("expected 10 OMCL, got exists=%t, qty=%f", exists, pos.Qty)
	}
}

