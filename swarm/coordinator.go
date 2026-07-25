package main

import (
	"context"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"pqr.info/mev/swarm/agents"
)

type Coordinator struct {
	mu           sync.RWMutex
	agents       map[string]agents.Agent
	activeMode   string // "paper", "shadow", "active"
	minEVWei     *big.Int
	latencyLimit float64
}

func NewCoordinator(minEV *big.Int, maxLatency float64, mode string) *Coordinator {
	return &Coordinator{
		agents:       make(map[string]agents.Agent),
		minEVWei:     minEV,
		latencyLimit: maxLatency,
		activeMode:   mode,
	}
}

func (c *Coordinator) RegisterAgent(agent agents.Agent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.agents[agent.ID()] = agent
	log.Info().Str("agent_id", agent.ID()).Msg("Swarm member registered successfully")
}

// ProcessStateIngestion runs analysis on all active agents for incoming market telemetry
func (c *Coordinator) ProcessStateIngestion(ctx context.Context, state agents.MarketState) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var proposals []agents.BundleProposal
	for _, agent := range c.agents {
		proposal, err := agent.Analyze(ctx, state)
		if err != nil {
			log.Error().Err(err).Str("agent", agent.ID()).Msg("Failed to analyze market state")
			continue
		}
		proposals = append(proposals, proposal)
	}

	bestProposal := c.EvaluateAndSelect(proposals)
	if bestProposal != nil {
		c.ExecuteProposal(*bestProposal)
	}
}

func (c *Coordinator) EvaluateAndSelect(props []agents.BundleProposal) *agents.BundleProposal {
	if len(props) == 0 {
		return nil
	}

	var best *agents.BundleProposal
	bestScore := big.NewInt(-1e18) // small initialization

	for i := range props {
		prop := props[i]

		// EV Constraint
		if prop.ExpectedValue.Cmp(c.minEVWei) < 0 {
			continue
		}

		// Latency Constraint
		if prop.LatencyMs > c.latencyLimit {
			continue
		}

		// Score computation: EV - GasCost - (Latency * scale)
		// Simulating Latency penalty: 1ms latency is worth ~0.0001 ETH (1e14 wei)
		latencyPenalty := big.NewInt(int64(prop.LatencyMs * 1e14))
		score := new(big.Int).Sub(prop.ExpectedValue, prop.GasCost)
		score = new(big.Int).Sub(score, latencyPenalty)

		if score.Cmp(bestScore) > 0 {
			bestScore = score
			best = &prop
		}
	}

	return best
}

func (c *Coordinator) ExecuteProposal(prop agents.BundleProposal) {
	switch c.activeMode {
	case "paper":
		log.Info().
			Str("mode", "paper-trading").
			Str("agent", prop.AgentID).
			Str("ev", prop.ExpectedValue.String()).
			Msg("📈 Shadow simulated transaction recorded for training log")
	case "shadow":
		log.Info().
			Str("mode", "shadow-mode").
			Str("agent", prop.AgentID).
			Str("ev", prop.ExpectedValue.String()).
			Msg("🔍 Live telemetry shadow event match detected")
	case "active":
		log.Warn().
			Str("mode", "active").
			Str("agent", prop.AgentID).
			Str("ev", prop.ExpectedValue.String()).
			Msg("🚨 Dispatching flashbots bundle to Arbitrum sequencer")
	}
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Initialize the Universal Z-Drive (Hetzner SFTP Mount)
	if err := InitZDriveMount(); err != nil {
		log.Warn().Err(err).Msg("Z-Drive failed to mount, some storage features may be unavailable")
	}

	coordinator := NewCoordinator(big.NewInt(10000000000000000), 2.5, "paper")

	// Start MEV bot orchestration (Task MEV-02)
	StartMEVOrchestrator()

	// Phase 122: Start Telemetry (Prometheus + Atlas Metrics API)
	metrics := NewMetrics()
	InitPrometheus()

	// Start Prometheus metrics endpoint asynchronously on 9090
	go func() {
		log.Info().Msg("Starting Prometheus metrics on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
			log.Error().Err(err).Msg("Prometheus server failed")
		}
	}()

	// Phase 119: Start Organ Atlas Mesh Reconciliation Loop
	atlas := NewOrganAtlas(metrics)
	nodesCfg := []NodeConfig{
		{Name: "MAX", Role: RoleInference, BaseURL: "http://192.168.12.234:8000"},
	}
	stopCh := make(chan struct{})
	go atlas.StartLoop(nodesCfg, 5*time.Second, stopCh)
	defer close(stopCh)

	maxExp, _ := new(big.Int).SetString("100000000000000000000", 10)
	agent := agents.NewArbitrumArbitrageAgent(
		"uni-camelot-arb",
		common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f352415231aa11"),
		common.HexToAddress("0x2bE450E63198e3b3206D767d02A24FBAf3cf0E26"),
		0.005,
		maxExp,
	)
	coordinator.RegisterAgent(agent)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simulate incoming Teleporter real-time market data
	log.Info().Msg("Starting teleporter mock feed stream...")
	for i := 0; i < 3; i++ {
		time.Sleep(1 * time.Second)
		state := agents.MarketState{
			GasPriceGwei: 28,
			BlockNumber:  1234567 + uint64(i),
			// Disparity gets larger, boosting EV
			UniV3Price:   big.NewInt(3000 + int64(i)),
			CamelotPrice: big.NewInt(3020 + int64(i*5)),
		}
		coordinator.ProcessStateIngestion(ctx, state)
	}

	// Phase 121: Start Atlas API (now injected with metrics)
	healingStore := NewHealingStore(nil)
	policyEngine := NewPolicyEngine(healingStore)
	var safetyEngine interface{}
	healer := NewHealerAgent(policyEngine, safetyEngine, healingStore)
	healer.Start()
	atlasAPI := NewAtlasAPI(atlas, metrics, healer)
	log.Info().Msg("Starting Atlas API on :8101")
	go func() {
		ouroboros := NewOuroborosMonitor()
	ouroboros.InstallLogHook()
	ouroboros.Start()

	defer ouroboros.Stop()

	if err := atlasAPI.Serve(":8101"); err != nil {
			log.Error().Err(err).Msg("Atlas API server failed")
		}
	}()

	// Phase 125: Start TimeMachineServer with 5D Router on :8081
	replayEngine := NewTimeMachineReplay(nil, nil)
	router5D := NewMeshPredictiveRouter5D(replayEngine)
	router5D.RegisterNode(PredictiveModelNode{
		NodeID:   "MAX",
		Address:  Address5D{X: 1, Y: 1, Z: 0, T: 0, W: 0},
		IsActive: true,
		Model:    "gemini-3.5-flash",
		Endpoint: "http://192.168.12.234:8081",
	})
	tmServer := NewTimeMachineServer(replayEngine, 8081)
	go func() {
		if err := tmServer.Start(); err != nil {
			log.Error().Err(err).Msg("TimeMachine server failed")
		}
	}()

	// Keep the organism running
	select {}
}
