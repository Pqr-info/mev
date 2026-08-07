package main

import "time"

type MeshConsensusDecision struct {
	DecisionID   string
	MutationID   string
	Approved     bool
	VotesFor     int
	VotesAgainst int
	Reason       string
	Timestamp    time.Time
}

type MutationVote struct {
	NodeID    string
	Mutation  MeshPolicyMutation
	Approve   bool
	Timestamp time.Time
}

type ConsensusEngine struct {
	quorumSize int
}

func NewConsensusEngine(quorumSize int) *ConsensusEngine {
	return &ConsensusEngine{
		quorumSize: quorumSize,
	}
}

func (c *ConsensusEngine) Decide(
	mutation MeshPolicyMutation,
	votes []MutationVote,
) MeshConsensusDecision {
	var forCount, againstCount int

	for _, v := range votes {
		if v.Approve {
			forCount++
		} else {
			againstCount++
		}
	}

	approved := forCount >= c.quorumSize && forCount > againstCount
	reason := "rejected"

	if approved {
		reason = "quorum-approved"
	}

	return MeshConsensusDecision{
		DecisionID:   "consensus-" + mutation.MutationID,
		MutationID:   mutation.MutationID,
		Approved:     approved,
		VotesFor:     forCount,
		VotesAgainst: againstCount,
		Reason:       reason,
		Timestamp:    time.Now(),
	}
}

func (c *ConsensusEngine) ApplyConsensus(
	policy *MeshPolicyConfig,
	mutation MeshPolicyMutation,
	decision MeshConsensusDecision,
) {
	if !decision.Approved {
		return
	}

	switch mutation.Field {
	case "risk-threshold":
		policy.RiskThreshold = mutation.NewValue
	}
}
