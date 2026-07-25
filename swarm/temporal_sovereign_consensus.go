package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"pqr.info/mev/time_machine"
)

type TemporalAction string

const (
	ActionRollback    TemporalAction = "ROLLBACK"
	ActionRollforward TemporalAction = "ROLLFORWARD"
	ActionPromote     TemporalAction = "PROMOTE"
)

type TemporalProposal struct {
	ProposalID string           `json:"proposal_id"`
	Node       string           `json:"node"`
	Expected   string           `json:"expected"`
	Actual     string           `json:"actual"`
	Role       NodeRole         `json:"role"`
	Timestamp  string           `json:"timestamp"`
	Options    []TemporalAction `json:"options"`
}

type Vote struct {
	Action     TemporalAction
	Confidence float64
	Reason     string
}

type SovereignConsensusEngine struct {
	atlas  *OrganAtlas
	client *http.Client
	policy *PolicyEngine
}

func NewSovereignConsensusEngine(atlas *OrganAtlas) *SovereignConsensusEngine {
	return &SovereignConsensusEngine{
		atlas: atlas,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		policy: NewPolicyEngine(NewHealingStore(nil)),
	}
}

func (ce *SovereignConsensusEngine) Resolve(proposal TemporalProposal, allNodes map[string]*NodeState) {
	start := time.Now()
	log.Info().Str("proposal", proposal.ProposalID).Str("node", proposal.Node).Msg("Initiating temporal consensus voting")

	if ce.atlas.metrics != nil {
		ce.atlas.metrics.mu.Lock()
		ce.atlas.metrics.Consensus.ProposalsTotal++
		ce.atlas.metrics.mu.Unlock()
	}

	// Filter voting nodes based on role scope
	votingNodes := ce.getVotingNodes(proposal.Role, allNodes)
	
	// Collect votes
	weightedVotes := make(map[TemporalAction]float64)
	totalWeight := 0.0

	for _, voter := range votingNodes {
		weight := ce.getNodeWeight(voter.Name, voter.Role)
		vote := ce.simulateVote(voter, proposal) // In a real system, we'd query the node via API

		weightedVotes[vote.Action] += weight * vote.Confidence
		totalWeight += weight

		log.Debug().
			Str("voter", voter.Name).
			Str("vote", string(vote.Action)).
			Float64("confidence", vote.Confidence).
			Float64("weight", weight).
			Msg("Vote received")
	}

	// Decision
	var decision TemporalAction
	var maxScore float64

	for action, score := range weightedVotes {
		if score > maxScore {
			maxScore = score
			decision = action
		}
	}

	confidence := 0.0
	if totalWeight > 0 {
		confidence = maxScore / totalWeight
	}

	log.Info().
		Str("proposal", proposal.ProposalID).
		Str("raw_decision", string(decision)).
		Float64("confidence", confidence).
		Msg("Consensus reached")

	latencyMs := float64(time.Since(start).Milliseconds())

	if ce.atlas.metrics != nil {
		ce.atlas.metrics.mu.Lock()
		ce.atlas.metrics.Consensus.DecisionsTotal++
		// Simple moving average for confidence
		ce.atlas.metrics.Consensus.AvgConfidence = (ce.atlas.metrics.Consensus.AvgConfidence + confidence) / 2
		ce.atlas.metrics.Consensus.AvgDecisionLatencyMs = (ce.atlas.metrics.Consensus.AvgDecisionLatencyMs + latencyMs) / 2
		consensusLatency.Set(ce.atlas.metrics.Consensus.AvgDecisionLatencyMs)
		ce.atlas.metrics.mu.Unlock()
	}

	// Calculate severity based on confidence
	var severity DriftSeverity = DriftMinor
	if confidence < 0.8 {
		severity = DriftMajor
	}
	if confidence < 0.6 {
		severity = DriftCritical
	}

	ctx := DriftContext{
		Node:       proposal.Node,
		Role:       proposal.Role,
		Severity:   severity,
		Metrics:    ce.atlas.metrics,
		Proposal:   proposal,
		Decision:   decision,
		Confidence: confidence,
	}

	finalDecision := ce.policy.Evaluate(ctx)

	log.Info().
		Str("proposal", proposal.ProposalID).
		Str("final_decision", string(finalDecision)).
		Msg("Policy engine evaluated consensus")

	// Emit Policy Decision to WAL
	brainDir := os.Getenv("GEMINI_BRAIN_DIR")
	if brainDir != "" {
		event := timemachine.MicrostructureEvent{
			"event_type":     "POLICY_DECISION",
			"node":           proposal.Node,
			"raw_decision":   string(decision),
			"final_decision": string(finalDecision),
			"severity":       string(severity),
			"timestamp":      time.Now().UTC().Format(time.RFC3339Nano),
		}
		_ = timemachine.WriteMicrostructureEvent(brainDir, time.Now(), event)
	}

	if finalDecision == "ESCALATE_TO_GEMINI" {
		log.Warn().Str("proposal", proposal.ProposalID).Msg("Policy Engine mandated escalation to Gemini/Gemma arbitration.")
		ce.escalateToGemini(proposal)
		return
	}

	ce.executeAction(finalDecision, proposal, allNodes[proposal.Node])
}

func (ce *SovereignConsensusEngine) getVotingNodes(driftRole NodeRole, allNodes map[string]*NodeState) []*NodeState {
	var voters []*NodeState
	for _, node := range allNodes {
		// Scoped by role mapping. E.g. Gateway drift is only voted on by Gateways.
		if node.Role == driftRole {
			voters = append(voters, node)
		} else if driftRole == RoleInference && node.Role == RoleControl {
			// Control might oversee inference, but following Copilot's strict scoping, 
			// let's keep them purely within their role cluster unless specifically extended.
			// Actually, let's just stick to role == driftRole for simplicity as defined in spec.
		}
	}
	return voters
}

func (ce *SovereignConsensusEngine) getNodeWeight(name string, role NodeRole) float64 {
	switch name {
	case "MAX", "CORE":
		return 1.0
	}

	switch role {
	case RoleInference, RoleControl:
		return 0.7
	case RoleGateway:
		return 0.3
	}
	return 0.1
}

func (ce *SovereignConsensusEngine) simulateVote(voter *NodeState, proposal TemporalProposal) Vote {
	// Simulated voting logic:
	// A canonical node expects others to match it.
	if voter.Name == "MAX" || voter.Name == "CORE" {
		if voter.StateHash == proposal.Expected {
			return Vote{Action: ActionRollback, Confidence: 0.95, Reason: "Canonical divergence; rollback required."}
		}
		// If canonical node has drifted itself? (It's usually the expected source, but just in case)
		return Vote{Action: ActionPromote, Confidence: 0.90, Reason: "Canonical node has evolved."}
	}

	// For peers, generally they want to follow canonical or rollback.
	return Vote{Action: ActionRollback, Confidence: 0.80, Reason: "Peer drift detected."}
}

func (ce *SovereignConsensusEngine) executeAction(action TemporalAction, proposal TemporalProposal, targetNode *NodeState) {
	brainDir := os.Getenv("GEMINI_BRAIN_DIR")
	
	switch action {
	case ActionRollback:
		ce.postAction(targetNode.Name, "rollback", proposal.Expected)
	case ActionRollforward:
		ce.postAction(targetNode.Name, "rollforward", proposal.Actual)
	case ActionPromote:
		ce.atlas.SetExpectedHash(proposal.Node, proposal.Actual)
	}

	event := timemachine.MicrostructureEvent{
		"event_type": "CONSENSUS_DECISION",
		"node":       proposal.Node,
		"action":     string(action),
		"old_hash":   proposal.Expected,
		"new_hash":   proposal.Actual,
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
	}
	
	if brainDir != "" {
		_ = timemachine.WriteMicrostructureEvent(brainDir, time.Now(), event)
	}
}

func (ce *SovereignConsensusEngine) postAction(nodeName string, action string, targetHash string) {
	log.Info().Str("node", nodeName).Str("action", action).Str("target", targetHash).Msg("Executing temporal POST request to node")
	// E.g., POST http://<NODE_IP>/time/rollback?target=expected_state_hash
	// For now, this is a simulated HTTP call since the node endpoints are assumed.
	url := fmt.Sprintf("http://%s/time/%s?target=%s", "localhost", action, targetHash)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte{}))
	_, err := ce.client.Do(req)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to execute temporal action via HTTP")
	}
}

func (ce *SovereignConsensusEngine) escalateToGemini(proposal TemporalProposal) {
	brainDir := os.Getenv("GEMINI_BRAIN_DIR")
	event := timemachine.MicrostructureEvent{
		"event_type": "ESCALATE_TO_GEMINI",
		"proposal":   proposal,
		"timestamp":  time.Now().UTC().Format(time.RFC3339Nano),
	}
	if brainDir != "" {
		_ = timemachine.WriteMicrostructureEvent(brainDir, time.Now(), event)
	}

	if ce.atlas.metrics != nil {
		ce.atlas.metrics.mu.Lock()
		ce.atlas.metrics.Consensus.ArbitrationTotal++
		arbitrationTotal.Inc()
		ce.atlas.metrics.mu.Unlock()
	}
}

func NewTemporalProposal(nodeName, expected, actual string, role NodeRole) TemporalProposal {
	return TemporalProposal{
		ProposalID: uuid.New().String(),
		Node:       nodeName,
		Expected:   expected,
		Actual:     actual,
		Role:       role,
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Options:    []TemporalAction{ActionRollback, ActionRollforward, ActionPromote},
	}
}
