package main

import (
	"time"
)

type ArbitrationReason string

const (
	ReasonPriority      ArbitrationReason = "PRIORITY"
	ReasonLineage       ArbitrationReason = "LINEAGE"
	ReasonGovernance    ArbitrationReason = "GOVERNANCE_WEIGHT"
	ReasonEmergency     ArbitrationReason = "EMERGENCY_OVERRIDE"
	ReasonTemporalScope ArbitrationReason = "TEMPORAL_SCOPE"
)

type ArbitrationResult struct {
	WinnerClause Clause
	LoserClauses []Clause
	Reason       string
	AtTime       time.Time
}

type Arbitrator interface {
	Arbitrate(
		domain Domain,
		actor ActorID,
		at time.Time,
		clauses []Clause,
	) (ArbitrationResult, error)
}

type defaultArbitrator struct{}

func NewDefaultArbitrator() Arbitrator {
	return &defaultArbitrator{}
}

func (a *defaultArbitrator) Arbitrate(
	domain Domain,
	actor ActorID,
	at time.Time,
	clauses []Clause,
) (ArbitrationResult, error) {
	if len(clauses) == 0 {
		return ArbitrationResult{
			Reason: "no clauses to arbitrate",
			AtTime: at,
		}, nil
	}

	winner := clauses[0]
	losers := make([]Clause, 0)

	for i := 1; i < len(clauses); i++ {
		c := clauses[i]
		if c.Priority > winner.Priority {
			losers = append(losers, winner)
			winner = c
		} else {
			losers = append(losers, c)
		}
	}

	return ArbitrationResult{
		WinnerClause: winner,
		LoserClauses: losers,
		Reason:       string(ReasonPriority),
		AtTime:       at,
	}, nil
}

type EngineWithArbitration struct {
	base       *TemporalSocialContractEngine
	arbitrator Arbitrator
}

func NewEngineWithArbitration(base *TemporalSocialContractEngine, arbitrator Arbitrator) *EngineWithArbitration {
	return &EngineWithArbitration{
		base:       base,
		arbitrator: arbitrator,
	}
}

func (e *EngineWithArbitration) ArbitrateClauses(
	domain Domain,
	actor ActorID,
	at time.Time,
	candidates []Clause,
) (ArbitrationResult, error) {
	return e.arbitrator.Arbitrate(domain, actor, at, candidates)
}
