package main

import (
	"testing"
	"time"
)

func TestTemporalGovernanceEnforcement(t *testing.T) {
	tg := NewTemporalGovernance()

	// Construct a timeline with cash persistently below CashFloor (1000) for 3 snapshots
	timeline := []PortfolioSnapshot{
		{
			Timestamp: time.Now().Add(-15 * time.Second),
			Portfolio: Portfolio{
				Cash: 500.0,
			},
		},
		{
			Timestamp: time.Now().Add(-10 * time.Second),
			Portfolio: Portfolio{
				Cash: 500.0,
			},
		},
		{
			Timestamp: time.Now().Add(-5 * time.Second),
			Portfolio: Portfolio{
				Cash: 500.0,
			},
		},
	}

	props := tg.Tick(timeline)
	if len(props) == 0 {
		t.Fatal("expected cash floor temporal governance proposal to be generated")
	}

	foundCashAdjust := false
	for _, prop := range props {
		if prop.Mutation.Kind == "cash_adjust" && prop.Mutation.Amount == 500.0 {
			foundCashAdjust = true
		}
	}

	if !foundCashAdjust {
		t.Error("expected cash_adjust proposal to restore floor to be generated")
	}
}
