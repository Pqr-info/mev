package main

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"

	"pqr.info/mev/swarm/agents"
)

type SwarmTopology struct {
	mu            sync.RWMutex
	ActiveAgents  map[string]agents.Agent
	PeerRelations map[string][]string // Maps AgentID to its peer strategy dependencies
}

func NewSwarmTopology() *SwarmTopology {
	return &SwarmTopology{
		ActiveAgents:  make(map[string]agents.Agent),
		PeerRelations: make(map[string][]string),
	}
}

func (st *SwarmTopology) AddAgent(agent agents.Agent, dependencies []string) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.ActiveAgents[agent.ID()] = agent
	st.PeerRelations[agent.ID()] = dependencies

	log.Info().
		Str("agent_id", agent.ID()).
		Interface("dependencies", dependencies).
		Msg("Swarm topology node registered")
}

func (st *SwarmTopology) GetAgent(id string) agents.Agent {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.ActiveAgents[id]
}

func (st *SwarmTopology) GetDependencies(id string) []string {
	st.mu.RLock()
	defer st.mu.RUnlock()
	return st.PeerRelations[id]
}

// IngestTelemetry runs the simulated flow over the registered topology
func (st *SwarmTopology) IngestTelemetry(ctx context.Context, state agents.MarketState) []agents.BundleProposal {
	st.mu.RLock()
	defer st.mu.RUnlock()

	var proposals []agents.BundleProposal
	for id, agent := range st.ActiveAgents {
		prop, err := agent.Analyze(ctx, state)
		if err != nil {
			log.Error().Err(err).Str("agent", id).Msg("Error running analysis in topology")
			continue
		}
		proposals = append(proposals, prop)
	}
	return proposals
}

// InitializeDefaultTopology configures the default active agents and their connections
func (st *SwarmTopology) InitializeDefaultTopology() {
	maxExp, _ := new(big.Int).SetString("100000000000000000000", 10)
	
	agentA := agents.NewArbitrumArbitrageAgent(
		"arb-uni-camelot",
		common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f352415231aa11"),
		common.HexToAddress("0x2bE450E63198e3b3206D767d02A24FBAf3cf0E26"),
		0.005,
		maxExp,
	)

	agentB := agents.NewArbitrumArbitrageAgent(
		"arb-sushi-camelot",
		common.HexToAddress("0x1b02dA8Cb0d097eB8D57A175b88c7D8b47997506"),
		common.HexToAddress("0x2bE450E63198e3b3206D767d02A24FBAf3cf0E26"),
		0.010,
		maxExp,
	)

	st.AddAgent(agentA, []string{})
	st.AddAgent(agentB, []string{"arb-uni-camelot"})
}
