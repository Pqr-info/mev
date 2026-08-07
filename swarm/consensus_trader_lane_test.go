package main

import (
	"testing"
)

func TestFinancialConsensusLane(t *testing.T) {
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

	engine := NewConsensusEngine(2) // quorum size 2

	// Scenario A: Validator rejects -> Immediate rejection
	mutForbidden := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "FORBIDDEN_COIN",
		Qty:       10.0,
		CostBasis: 10.0,
	}
	decision := engine.DecideFinancial(p, mutForbidden, config, nil)
	if decision.Approved {
		t.Error("expected forbidden instrument to be rejected by the consensus lane")
	}

	// Scenario B: Validator passes, but quorum fails
	mutValid := FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "OMCL",
		Qty:       10.0,
		CostBasis: 50.0,
	}
	votesFailing := []MutationVote{
		{NodeID: "node1", Approve: true},
		{NodeID: "node2", Approve: false},
	}
	decision = engine.DecideFinancial(p, mutValid, config, votesFailing)
	if decision.Approved {
		t.Error("expected consensus to fail due to insufficient quorum votes")
	}

	// Scenario C: Validator passes and quorum approves
	votesPassing := []MutationVote{
		{NodeID: "node1", Approve: true},
		{NodeID: "node2", Approve: true},
	}
	decision = engine.DecideFinancial(p, mutValid, config, votesPassing)
	if !decision.Approved {
		t.Error("expected consensus to approve the valid transaction with passing votes")
	}
}
