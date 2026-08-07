package main

import "time"

type TemporalConstitution struct {
	ReplayLimits          ReplayLimits
	TimeslipRules         TimeslipRules
	MutationBoundaries    MutationBoundaries
	FederationObligations FederationObligations
	QuorumRules           QuorumRules
	EmergencyPowers       EmergencyPowers
}

type ReplayLimits struct {
	MaxReplayPerEpoch int
	AllowFractured    bool
}

type TimeslipRules struct {
	AllowTimeslips       bool
	MaxTimeslipDeviation float64
}

type MutationBoundaries struct {
	MaxMutationRate  float64
	RequireConsensus bool
}

type FederationObligations struct {
	MinSyncFrequencySeconds int
	DivergenceEscalation    bool
}

type QuorumRules struct {
	RequiredVotes int
	Supermajority bool
}

type EmergencyPowers struct {
	EnableDuringUnstable  bool
	EnableDuringFractured bool
	ClampMutationRate     float64
}

type ConstitutionEngine struct {
	constitution TemporalConstitution
}

func NewConstitutionEngine(c TemporalConstitution) *ConstitutionEngine {
	return &ConstitutionEngine{
		constitution: c,
	}
}

func (ce *ConstitutionEngine) Interpret(
	directive TemporalGovernanceDirective,
) TemporalGovernanceDirective {

	if !ce.constitution.ReplayLimits.AllowFractured &&
		directive.Reason == "fractured-epoch" {
		directive.ReplayAllowed = false
	}

	if !ce.constitution.TimeslipRules.AllowTimeslips {
		directive.TimeslipAllowed = false
	}

	if directive.MutationRateCap > ce.constitution.MutationBoundaries.MaxMutationRate {
		directive.MutationRateCap = ce.constitution.MutationBoundaries.MaxMutationRate
	}

	minSync := time.Duration(ce.constitution.FederationObligations.MinSyncFrequencySeconds) * time.Second
	if directive.FederationSyncFreq < minSync {
		directive.FederationSyncFreq = minSync
	}

	if directive.Reason == "unstable-epoch" &&
		ce.constitution.EmergencyPowers.EnableDuringUnstable {
		directive.MutationRateCap = ce.constitution.EmergencyPowers.ClampMutationRate
	}
	if directive.Reason == "fractured-epoch" &&
		ce.constitution.EmergencyPowers.EnableDuringFractured {
		directive.MutationRateCap = ce.constitution.EmergencyPowers.ClampMutationRate
	}

	return directive
}
