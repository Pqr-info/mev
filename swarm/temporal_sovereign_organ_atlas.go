package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"pqr.info/mev/time_machine"

	"github.com/rs/zerolog/log"
)

type NodeRole string

const (
	RoleInference NodeRole = "INFERENCE"
	RoleControl   NodeRole = "CONTROL"
	RoleGateway   NodeRole = "GATEWAY"
)

type NodeConfig struct {
	Name    string
	Role    NodeRole
	BaseURL string
}

type NodeState struct {
	Name      string
	Role      NodeRole
	StateHash string
	Health    string
	Manifest  *timemachine.StateManifest
}

// OrganAtlas coordinates the state across all nodes.
type OrganAtlas struct {
	mu       sync.RWMutex
	expected map[string]string         // nodeName -> expected STATE_HASH
	nodes    map[string]*NodeState     // nodeName -> NodeState
	client   *http.Client
	engine   *SovereignConsensusEngine
	metrics  *Metrics
}

func NewOrganAtlas(metrics *Metrics) *OrganAtlas {
	oa := &OrganAtlas{
		expected: make(map[string]string),
		nodes:    make(map[string]*NodeState),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		metrics: metrics,
	}
	oa.engine = NewSovereignConsensusEngine(oa)
	return oa
}

func (oa *OrganAtlas) SetExpectedHash(name, hash string) {
	oa.mu.Lock()
	defer oa.mu.Unlock()
	oa.expected[name] = hash
}

func (oa *OrganAtlas) pollNode(name, baseURL string) (*NodeState, error) {
	url := fmt.Sprintf("%s/state/manifest", baseURL)
	resp, err := oa.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", resp.StatusCode)
	}

	var manifest timemachine.StateManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	hash, err := manifest.StateHash()
	if err != nil {
		return nil, fmt.Errorf("hash: %w", err)
	}

	// derive Health from manifest (services, models, brain)
	health := "OK"
	for _, svc := range manifest.Services {
		if svc.Status != "running" && svc.Status != "ok" {
			health = "DEGRADED"
		}
	}

	return &NodeState{
		Name:      name,
		StateHash: hash,
		Health:    health,
		Manifest:  &manifest,
	}, nil
}

func (oa *OrganAtlas) ReconcileAll(nodesCfg []NodeConfig) {
	oa.mu.Lock()
	defer oa.mu.Unlock()

	brainDir := os.Getenv("GEMINI_BRAIN_DIR")

	for _, n := range nodesCfg {
		st, err := oa.pollNode(n.Name, n.BaseURL)
		if err != nil {
			log.Warn().Str("node", n.Name).Err(err).Msg("Failed to poll node state")
			st = &NodeState{
				Name:   n.Name,
				Role:   n.Role,
				Health: "UNREACHABLE",
			}
			oa.nodes[n.Name] = st

			if oa.metrics != nil {
				oa.metrics.mu.Lock()
				oa.metrics.Health.NodeReachabilityFailures++
				oa.metrics.mu.Unlock()
			}

			continue
		}
		
		st.Role = n.Role
		oa.nodes[n.Name] = st

		exp := oa.expected[n.Name]
		if exp != "" && exp != st.StateHash {
			if oa.metrics != nil {
				oa.metrics.mu.Lock()
				oa.metrics.Temporal.DriftEventsTotal++
				driftEvents.Inc()
				oa.metrics.mu.Unlock()
			}

			emitDriftEvent(brainDir, n.Name, exp, st.StateHash, string(n.Role))
			
			// Phase 120: Initiate Temporal Consensus
			proposal := NewTemporalProposal(n.Name, exp, st.StateHash, n.Role)
			oa.engine.Resolve(proposal, oa.nodes)
		}
	}
}

func emitDriftEvent(brainDir, nodeName, expectedHash, actualHash, role string) {
	event := timemachine.MicrostructureEvent{
		"type":                "STATE_DRIFT",
		"node":                nodeName,
		"expected_state_hash": expectedHash,
		"actual_state_hash":   actualHash,
		"role":                role,
		"timestamp":           time.Now().UTC().Format(time.RFC3339Nano),
	}

	if brainDir == "" {
		log.Warn().Msg("GEMINI_BRAIN_DIR not set; cannot emit STATE_DRIFT to WAL")
		return
	}

	err := timemachine.WriteMicrostructureEvent(brainDir, time.Now(), event)
	if err != nil {
		log.Error().Err(err).Str("node", nodeName).Msg("Failed to emit STATE_DRIFT to time machine WAL")
	} else {
		log.Info().Str("node", nodeName).Msg("Emitted STATE_DRIFT event to WAL successfully")
	}
}

func (oa *OrganAtlas) StartLoop(nodesCfg []NodeConfig, interval time.Duration, stopCh <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			oa.ReconcileAll(nodesCfg)
		}
	}
}
