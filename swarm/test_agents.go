package main

import (
	"math/rand"
	"time"

	"github.com/rs/zerolog/log"
)

// -----------------------------
// Agent Interface
// -----------------------------

type SwarmAgent interface {
	ID() string
	Analyze(state *SimulatedState) AgentAnalysis
	Execute(analysis AgentAnalysis) AgentResult
}

// -----------------------------
// AgentAnalysis + AgentResult
// -----------------------------

type AgentAnalysis struct {
	AgentID          string
	OpportunityScore float64
	GasEstimate      float64
	Confidence       float64
	Notes            string
}

type AgentResult struct {
	Executed bool
	Profit   float64
	Latency  time.Duration
	Notes    string
}

// -----------------------------
// Simulated State
// -----------------------------

type SimulatedState struct {
	BlockNumber int
	PoolDelta   float64
	GasPrice    float64
	Volatility  float64
}

// -----------------------------
// Test Agent Implementation
// -----------------------------

type TestAgent struct {
	id string
}

func NewTestAgent(id string) *TestAgent {
	return &TestAgent{id: id}
}

func (a *TestAgent) ID() string {
	return a.id
}

func (a *TestAgent) Analyze(state *SimulatedState) AgentAnalysis {
	oppScore := rand.Float64() * state.PoolDelta
	gasEst := state.GasPrice * (0.8 + rand.Float64()*0.4)
	confidence := 0.90
	if state.Volatility > 0.6 {
		confidence = 0.60
	}
	return AgentAnalysis{
		AgentID:          a.id,
		OpportunityScore: oppScore,
		GasEstimate:      gasEst,
		Confidence:       confidence,
		Notes:            "Mock analysis completed",
	}
}

func (a *TestAgent) Execute(analysis AgentAnalysis) AgentResult {
	if analysis.Confidence < 0.70 {
		return AgentResult{
			Executed: false,
			Profit:   0.0,
			Latency:  0,
			Notes:    "Execution skipped: low confidence",
		}
	}
	profit := analysis.OpportunityScore * 10.0
	latency := time.Duration(10+rand.Intn(40)) * time.Millisecond
	return AgentResult{
		Executed: true,
		Profit:   profit,
		Latency:  latency,
		Notes:    "Execution simulation complete",
	}
}

// -----------------------------
// Swarm Manager
// -----------------------------

type TestSwarmManager struct {
	agents []SwarmAgent
}

func NewTestSwarmManager() *TestSwarmManager {
	return &TestSwarmManager{
		agents: make([]SwarmAgent, 0),
	}
}

func (m *TestSwarmManager) AddAgent(agent SwarmAgent) {
	m.agents = append(m.agents, agent)
	log.Info().Str("agent_id", agent.ID()).Msg("Swarm test agent registered")
}

func (m *TestSwarmManager) RunSimulationStep(state *SimulatedState) {
	log.Info().Int("block", state.BlockNumber).Msg("Running swarm step simulation")
	for _, agent := range m.agents {
		analysis := agent.Analyze(state)
		result := agent.Execute(analysis)
		if result.Executed {
			log.Info().
				Str("agent", agent.ID()).
				Float64("profit", result.Profit).
				Str("latency", result.Latency.String()).
				Msg("Agent executed trade successfully")
		}
	}
}
