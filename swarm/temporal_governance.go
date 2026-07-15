package main

import "time"

type TemporalGovernanceDirective struct {
	DirectiveID        string
	Epoch              string
	ReplayAllowed      bool
	TimeslipAllowed    bool
	MutationRateCap    float64
	FederationSyncFreq time.Duration
	EscalateToCouncil  bool
	Reason             string
	Timestamp          time.Time
}

type TemporalGovernanceEngine struct {
	tso        *TemporalStabilityOracle
	federation *TemporalFederationEngine
	consensus  *ConsensusEngine
}

func NewTemporalGovernanceEngine(
	tso *TemporalStabilityOracle,
	fed *TemporalFederationEngine,
	cons *ConsensusEngine,
) *TemporalGovernanceEngine {
	return &TemporalGovernanceEngine{
		tso:        tso,
		federation: fed,
		consensus:  cons,
	}
}

func (g *TemporalGovernanceEngine) Govern(
	epoch string,
	nodeID string,
	mutation MeshPolicyMutation,
	votes []MutationVote,
	remote FederationPacket,
) TemporalGovernanceDirective {
	stability := g.tso.Evaluate(epoch)
	divergence := g.federation.Compare(remote)
	decision := g.consensus.Decide(mutation, votes)

	replayAllowed := true
	timeslipAllowed := true
	mutationRateCap := 1.0
	syncFreq := time.Minute
	escalate := false
	reason := "normal-governance"

	if stability.StabilityClass == "fractured" {
		replayAllowed = false
		timeslipAllowed = false
		mutationRateCap = 0.3
		syncFreq = 10 * time.Second
		escalate = true
		reason = "fractured-epoch"
	} else if stability.StabilityClass == "unstable" {
		timeslipAllowed = true
		mutationRateCap = 0.6
		syncFreq = 30 * time.Second
		reason = "unstable-epoch"
	}

	if divergence > 0.5 {
		mutationRateCap *= 0.8
		syncFreq = 15 * time.Second
		escalate = true
		reason = "federation-divergence"
	}

	if !decision.Approved {
		mutationRateCap = 0.0
		reason = "consensus-rejected-mutation"
	}

	return TemporalGovernanceDirective{
		DirectiveID:        "gov-" + epoch + "-" + nodeID,
		Epoch:              epoch,
		ReplayAllowed:      replayAllowed,
		TimeslipAllowed:    timeslipAllowed,
		MutationRateCap:    mutationRateCap,
		FederationSyncFreq: syncFreq,
		EscalateToCouncil:  escalate,
		Reason:             reason,
		Timestamp:          time.Now(),
	}
}
