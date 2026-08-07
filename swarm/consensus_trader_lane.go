package main

import (
	"fmt"
	"time"
)

// DecideFinancial evaluates a FinancialMutation under Article II validator rules and returns a constitutional decision
func (c *ConsensusEngine) DecideFinancial(
	portfolio *Portfolio,
	mutation FinancialMutation,
	config ValidatorConfig,
	votes []MutationVote,
) MeshConsensusDecision {
	// 1. Run the Article II Constitutional Validator
	allowed, reason := ValidateFinancialMutation(portfolio, mutation, config)
	if !allowed {
		return MeshConsensusDecision{
			DecisionID:   "consensus-financial-invalid",
			MutationID:   string(mutation.Kind),
			Approved:     false,
			VotesFor:     0,
			VotesAgainst: len(votes),
			Reason:       fmt.Sprintf("Article II Constitutional Rejection: %s", reason),
			Timestamp:    time.Now(),
		}
	}

	// 2. Process quorum votes if constitutional checks pass
	var forCount, againstCount int
	for _, v := range votes {
		if v.Approve {
			forCount++
		} else {
			againstCount++
		}
	}

	// Default to approved if no votes are provided but validator passed, or check quorum size
	approved := true
	decisionReason := "allowed-by-validator-default"

	if len(votes) > 0 {
		approved = forCount >= c.quorumSize && forCount > againstCount
		if approved {
			decisionReason = "quorum-approved"
		} else {
			decisionReason = "quorum-rejected"
		}
	}

	return MeshConsensusDecision{
		DecisionID:   "consensus-financial-" + string(mutation.Kind) + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		MutationID:   string(mutation.Kind),
		Approved:     approved,
		VotesFor:     forCount,
		VotesAgainst: againstCount,
		Reason:       decisionReason,
		Timestamp:    time.Now(),
	}
}
