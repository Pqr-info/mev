package main

import (
	"testing"
	"time"
)

func TestMotorTraderHeuristics(t *testing.T) {
	pe := NewTemporalPortfolioEngine()
	ce := NewConsensusEngine(1)
	bridge := NewTraderBridge(pe, ce)
	mt := NewMotorTrader(pe, ce, bridge)

	config := DefaultValidatorConfig()

	pe.GetOrCreate("test-motor")

	// Initial cash injection
	pe.ApplyMutation("test-motor", FinancialMutation{
		Kind:   CashAdjust,
		Amount: 5000.0,
	})

	// Add an overly large position (exceeding mt.MaxExposure of 5000)
	pe.ApplyMutation("test-motor", FinancialMutation{
		Kind:      PortfolioAdd,
		Symbol:    "TSM",
		Qty:       20.0,
		CostBasis: 400.0, // Exposure = 8000
	})

	props := mt.Tick("test-motor", config)
	if len(props) == 0 {
		t.Fatal("expected exposure drift rebalance proposal to be generated")
	}

	foundRebalance := false
	for _, prop := range props {
		if prop.Mutation.Symbol == "TSM" && prop.Mutation.Kind == string(PortfolioRemove) {
			foundRebalance = true
			if prop.Mutation.Qty != 7.5 { // (8000 - 5000) / 400 = 7.5 qty to remove
				t.Errorf("expected qty reduction of 7.5, got %f", prop.Mutation.Qty)
			}
		}
	}

	if !foundRebalance {
		t.Error("TSM exposure drift rebalance proposal not found")
	}
}

func TestMotorTraderTemporalFeedback(t *testing.T) {
	pe := NewTemporalPortfolioEngine()
	ce := NewConsensusEngine(1)
	bridge := NewTraderBridge(pe, ce)
	mt := NewMotorTrader(pe, ce, bridge)

	// Construct a timeline showing leverage oscillation (e.g. 2.5 -> 0.5 -> 2.5)
	timeline := []PortfolioSnapshot{
		{
			Timestamp: time.Now().Add(-10 * time.Second),
			LineageID: "l1",
			Portfolio: Portfolio{
				Risk: RiskProfile{CurrentLeverage: 2.5},
			},
		},
		{
			Timestamp: time.Now().Add(-5 * time.Second),
			LineageID: "l2",
			Portfolio: Portfolio{
				Risk: RiskProfile{CurrentLeverage: 0.5},
			},
		},
		{
			Timestamp: time.Now(),
			LineageID: "l3",
			Portfolio: Portfolio{
				Risk: RiskProfile{CurrentLeverage: 2.5},
			},
		},
	}

	props := mt.TemporalTick(timeline)
	if len(props) == 0 {
		t.Fatal("expected temporal stabilization proposal to be generated")
	}

	if props[0].Mutation.Kind != string(RiskProfileUpdate) {
		t.Errorf("expected risk profile update mutation, got %s", props[0].Mutation.Kind)
	}
}
