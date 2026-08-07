package main

import "time"

type MeshPolicyMutation struct {
	MutationID string
	Field      string
	OldValue   float64
	NewValue   float64
	Reason     string
	Epoch      string
	Timestamp  time.Time
}

type MeshPolicyConfig struct {
	RiskThreshold float64
}

type PolicyMutationEngine struct {
	tso       *TemporalStabilityOracle
	council   *EvolutionCouncil
	bicameral *BicameralEngine
	validator *OracleReplayValidator
}

func NewPolicyMutationEngine(
	tso *TemporalStabilityOracle,
	council *EvolutionCouncil,
	bic *BicameralEngine,
	val *OracleReplayValidator,
) *PolicyMutationEngine {
	return &PolicyMutationEngine{
		tso:       tso,
		council:   council,
		bicameral: bic,
		validator: val,
	}
}

func (p *PolicyMutationEngine) Mutate(epoch string) MeshPolicyMutation {
	stability := p.tso.Evaluate(epoch)
	insight := p.council.FormInsight()
	decision := p.bicameral.Resolve(MeshQuorum{}) // Using empty mock quorum
	consistency := p.validator.Validate(epoch, []TemporalEvent{})

	field := "risk-threshold"
	old := insight.RiskThreshold
	new := old
	reason := "no-change"

	if stability.StabilityClass == "unstable" {
		new = old * 0.9
		reason = "unstable-epoch"
	}

	if stability.StabilityClass == "fractured" {
		new = old * 0.7
		reason = "fractured-epoch"
	}

	if consistency > 0.5 {
		new = old * 1.1
		reason = "oracle-divergence"
	}

	if !decision.FinalDecision {
		new = old
		reason = "bicameral-rejected-mutation"
	}

	return MeshPolicyMutation{
		MutationID: "mutation-" + epoch,
		Field:      field,
		OldValue:   old,
		NewValue:   new,
		Reason:     reason,
		Epoch:      epoch,
		Timestamp:  time.Now(),
	}
}

func (p *PolicyMutationEngine) ApplyMutation(policy *MeshPolicyConfig, m MeshPolicyMutation) {
	switch m.Field {
	case "risk-threshold":
		policy.RiskThreshold = m.NewValue
	}
}
