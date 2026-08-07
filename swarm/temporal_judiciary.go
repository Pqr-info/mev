package main

import "time"

type TemporalRuling struct {
	RulingID  string
	Epoch     string
	NodeID    string
	Violation string
	Verdict   string
	Directive TemporalGovernanceDirective
	Timestamp time.Time
}

type TemporalJudiciaryEngine struct {
	constitution TemporalConstitution
}

func NewTemporalJudiciaryEngine(c TemporalConstitution) *TemporalJudiciaryEngine {
	return &TemporalJudiciaryEngine{
		constitution: c,
	}
}

func (j *TemporalJudiciaryEngine) Adjudicate(
	epoch string,
	nodeID string,
	directive TemporalGovernanceDirective,
) TemporalRuling {

	violation := ""
	verdict := "upheld"

	// replay violations
	if !j.constitution.ReplayLimits.AllowFractured &&
		directive.Reason == "fractured-epoch" &&
		directive.ReplayAllowed {
		violation = "replay-violation"
		directive.ReplayAllowed = false
		verdict = "corrected"
	}

	// timeslip violations
	if !j.constitution.TimeslipRules.AllowTimeslips &&
		directive.TimeslipAllowed {
		violation = "timeslip-violation"
		directive.TimeslipAllowed = false
		verdict = "corrected"
	}

	// mutation boundary violations
	if directive.MutationRateCap > j.constitution.MutationBoundaries.MaxMutationRate {
		violation = "mutation-boundary-violation"
		directive.MutationRateCap = j.constitution.MutationBoundaries.MaxMutationRate
		verdict = "corrected"
	}

	// federation obligations
	minSync := time.Duration(j.constitution.FederationObligations.MinSyncFrequencySeconds) * time.Second
	if directive.FederationSyncFreq < minSync {
		violation = "federation-obligation-violation"
		directive.FederationSyncFreq = minSync
		verdict = "corrected"
	}

	return TemporalRuling{
		RulingID:  "ruling-" + epoch + "-" + nodeID,
		Epoch:     epoch,
		NodeID:    nodeID,
		Violation: violation,
		Verdict:   verdict,
		Directive: directive,
		Timestamp: time.Now(),
	}
}
