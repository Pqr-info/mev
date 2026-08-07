package governance

import (
	"testing"
	"time"
)

func TestConstitutionAndAdjudication(t *testing.T) {
	c := TemporalConstitution{
		ReplayLimits: ReplayLimits{
			MaxReplayPerEpoch: 5,
			AllowFractured:    false,
		},
		TimeslipRules: TimeslipRules{
			AllowTimeslips: false,
		},
		MutationBoundaries: MutationBoundaries{
			MaxMutationRate: 0.15,
		},
		FederationObligations: FederationObligations{
			MinSyncFrequencySeconds: 10,
		},
	}

	ce := NewConstitutionEngine(c)
	je := NewTemporalJudiciaryEngine(ce)

	// Valid directive
	d := TemporalGovernanceDirective{
		Reason:             "normal",
		ReplayAllowed:      true,
		TimeslipAllowed:    false,
		MutationRateCap:    0.10,
		FederationSyncFreq: 15 * time.Second,
	}

	res := ce.Interpret(d)
	if res.MutationRateCap != 0.10 {
		t.Errorf("expected 0.10 mutation cap, got %v", res.MutationRateCap)
	}

	// Invalid directive (violation)
	dBad := TemporalGovernanceDirective{
		Reason:             "fractured-epoch",
		ReplayAllowed:      true,
		TimeslipAllowed:    true,
		MutationRateCap:    0.50,
		FederationSyncFreq: 5 * time.Second,
	}

	ruling := je.Adjudicate("epoch-1", "node-1", dBad)
	if ruling.Verdict != "corrected" {
		t.Errorf("expected verdict corrected, got %s", ruling.Verdict)
	}
	if ruling.Directive.ReplayAllowed {
		t.Error("expected replay to be corrected/disabled")
	}
	if ruling.Directive.MutationRateCap > 0.15 {
		t.Errorf("expected mutation rate to be clamped to 0.15, got %v", ruling.Directive.MutationRateCap)
	}
}

func TestLegislatureAmendments(t *testing.T) {
	c := TemporalConstitution{
		MutationBoundaries: MutationBoundaries{
			MaxMutationRate: 0.15,
		},
	}

	ce := NewConstitutionEngine(c)
	leg := NewTemporalLegislature(ce)

	// Valid amendment
	err := leg.ApplyConstitutionAmendment(Amendment{
		Field:    "MaxMutationRate",
		NewValue: 0.25,
	})
	if err != nil {
		t.Fatalf("failed to apply amendment: %v", err)
	}

	if ce.GetConstitution().MutationBoundaries.MaxMutationRate != 0.25 {
		t.Errorf("expected rate to be amended to 0.25, got %v", ce.GetConstitution().MutationBoundaries.MaxMutationRate)
	}

	// Invalid type amendment (should error, not panic)
	err = leg.ApplyConstitutionAmendment(Amendment{
		Field:    "MaxMutationRate",
		NewValue: "not-a-float",
	})
	if err == nil {
		t.Error("expected error on invalid type amendment")
	}
}

func TestArbitrationDeterministicTieBreaker(t *testing.T) {
	ae := NewArbitrationEngine()

	clauses := []Clause{
		{Name: "ClauseB", Priority: 10, Rule: "Rule B"},
		{Name: "ClauseA", Priority: 10, Rule: "Rule A"},
		{Name: "ClauseC", Priority: 5, Rule: "Rule C"},
	}

	winner, losers, err := ae.Arbitrate(clauses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Priority tie-breaker: ClauseA should win over ClauseB because A < B lexicographically
	if winner.Name != "ClauseA" {
		t.Errorf("expected winner ClauseA, got %s", winner.Name)
	}
	if len(losers) != 2 {
		t.Errorf("expected 2 losers, got %d", len(losers))
	}
}

func TestSanctionsAndAppealsInterlock(t *testing.T) {
	engine := NewSanctionsAndAppealsEngine()

	// Direct sanction
	s := &Sanction{
		SanctionID: "s1",
		Actor:      "node-1",
		Severity:   8,
	}

	engine.EnforceSanction(s)
	if s.Status != "ACTIVE" {
		t.Errorf("expected active sanction, got %s", s.Status)
	}

	// Submit appeal should quarantine sanction to PENDING_APPEAL status
	a := &Appeal{
		AppealID:  "a1",
		Actor:     "node-1",
		Reason:    "faulty telemetry",
		Status:    "SUBMITTED",
		Timestamp: time.Now(),
	}

	engine.SubmitAppeal(a)

	engine.mu.RLock()
	sStatus := s.Status
	engine.mu.RUnlock()

	if sStatus != "PENDING_APPEAL" {
		t.Errorf("expected sanction quarantined to PENDING_APPEAL, got %s", sStatus)
	}

	// New sanction while appeal is active should be created as PENDING_APPEAL immediately
	s2 := &Sanction{
		SanctionID: "s2",
		Actor:      "node-1",
		Severity:   6,
	}
	engine.EnforceSanction(s2)
	if s2.Status != "PENDING_APPEAL" {
		t.Errorf("expected second sanction created as PENDING_APPEAL immediately, got %s", s2.Status)
	}
}
