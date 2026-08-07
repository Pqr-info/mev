package main

import (
	"time"
)

// -----------------------------
// Evolution Council Mockup
// -----------------------------

type Insight struct {
	Risk          float64
	RiskThreshold float64
}

type EvolutionCouncil struct {
	baseRisk float64
}

func NewEvolutionCouncil(risk float64) *EvolutionCouncil {
	return &EvolutionCouncil{
		baseRisk: risk,
	}
}

func (c *EvolutionCouncil) FormInsight() Insight {
	return Insight{
		Risk:          c.baseRisk,
		RiskThreshold: 0.5,
	}
}

// -----------------------------
// Bicameral Decision Model
// -----------------------------

type BicameralDecision struct {
	QuorumApproved  bool
	CouncilApproved bool
	FinalDecision   bool
	Reason          string
	Timestamp       time.Time
}

// -----------------------------
// Bicameral Engine
// -----------------------------

type BicameralEngine struct {
	council *EvolutionCouncil
}

func NewBicameralEngine(c *EvolutionCouncil) *BicameralEngine {
	return &BicameralEngine{
		council: c,
	}
}

// -----------------------------
// Bicameral Resolution Logic
// -----------------------------

func (b *BicameralEngine) Resolve(quorum MeshQuorum) BicameralDecision {
	insight := b.council.FormInsight()

	quorumApproved := quorum.Formed
	councilApproved := insight.Risk < insight.RiskThreshold

	final := quorumApproved && councilApproved

	reason := "both chambers approved"
	if !quorumApproved && councilApproved {
		reason = "quorum rejected"
	}
	if quorumApproved && !councilApproved {
		reason = "council rejected"
	}
	if !quorumApproved && !councilApproved {
		reason = "both chambers rejected"
	}

	return BicameralDecision{
		QuorumApproved:  quorumApproved,
		CouncilApproved: councilApproved,
		FinalDecision:   final,
		Reason:          reason,
		Timestamp:       time.Now(),
	}
}
