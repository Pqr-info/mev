package main

import (
	"context"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
)

type CoordinatorDecision struct {
	AgentID    string
	Score      float64
	Confidence float64
	Timestamp  time.Time
	Notes      string
}

type NewSwarmCoordinator struct {
	agents []SwarmAgent
	router TeleporterRouter
}

func NewNewSwarmCoordinator(router TeleporterRouter) *NewSwarmCoordinator {
	return &NewSwarmCoordinator{
		agents: make([]SwarmAgent, 0),
		router: router,
	}
}

func (c *NewSwarmCoordinator) RegisterAgent(a SwarmAgent) {
	c.agents = append(c.agents, a)
}

func (c *NewSwarmCoordinator) Tick(ctx context.Context, state *SimulatedState) *CoordinatorDecision {
	type candidate struct {
		agent    SwarmAgent
		analysis AgentAnalysis
	}

	var candidates []candidate

	// Gather analyses from all registered agents
	for _, agent := range c.agents {
		analysis := agent.Analyze(state)
		candidates = append(candidates, candidate{
			agent:    agent,
			analysis: analysis,
		})
	}

	if len(candidates) == 0 {
		return nil
	}

	// Rank based on score (OpportunityScore)
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].analysis.OpportunityScore > candidates[j].analysis.OpportunityScore
	})

	best := candidates[0]

	// Execute best candidate
	result := best.agent.Execute(best.analysis)
	if result.Executed {
		log.Info().
			Str("agent", best.agent.ID()).
			Float64("score", best.analysis.OpportunityScore).
			Float64("profit", result.Profit).
			Msg("Swarm Coordinator selected and executed best candidate opportunity")

		// Emit synthetic telemetry into the router
		env := &FirehoseEnvelope{
			Source:      "swarm-coordinator",
			StreamID:    "simulated-production",
			Timestamp:   time.Now(),
			PayloadType: "mesh_event",
		}
		payload := MeshEventPayload{
			SigmaID:    best.agent.ID(),
			Agent:      "test-agent",
			Files:      []string{},
			RiskScore:  best.analysis.OpportunityScore,
			Confidence: best.analysis.Confidence,
		}

		_ = c.router.Route(ctx, env, payload)

		return &CoordinatorDecision{
			AgentID:    best.agent.ID(),
			Score:      best.analysis.OpportunityScore,
			Confidence: best.analysis.Confidence,
			Timestamp:  time.Now(),
			Notes:      result.Notes,
		}
	}

	return nil
}
