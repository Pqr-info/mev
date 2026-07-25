package governance

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"pqr.info/jetweb-core/core/go/pikr"
)

// ==========================================
// 1. Swarm Supervisor Types and Policies
// ==========================================

type AgentStatus string

const (
	StatusIdle    AgentStatus = "IDLE"
	StatusRunning AgentStatus = "RUNNING"
	StatusError   AgentStatus = "ERROR"
)

type AgentSpec struct {
	Name        string
	Role        string
	Lineage     string
	MaxRetries  int
	CommandPath string
}

type AgentInstance struct {
	Spec     AgentSpec
	Status   AgentStatus
	Identity *pikr.Identity5
	Errors   int
	LastSeen time.Time
	PID      int
}

type PolicyConfig struct {
	Version      string              `json:"version"`
	LineageRules map[string][]string `json:"lineage_rules"`
}

type PolicyEngine interface {
	AllowsIdentity(id *pikr.Identity5, role string) error
	AllowsLineage(lineage string, role string) error
	UpdateFromConfig(cfg PolicyConfig) error
}

type DefaultPolicyEngine struct {
	mu              sync.RWMutex
	AllowedLineages map[string][]string
}

func NewDefaultPolicyEngine() PolicyEngine {
	return &DefaultPolicyEngine{
		AllowedLineages: map[string][]string{
			"Sentinel": {"corridor.east", "corridor.west", "corridor.central"},
			"Analyst":  {"corridor.east", "corridor.west", "corridor.central"},
		},
	}
}

func (pe *DefaultPolicyEngine) AllowsIdentity(id *pikr.Identity5, role string) error {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	sk := id.RecoveryMatrix().SovereignKey()
	if len(sk) != 32 {
		return fmt.Errorf("cryptographic identity verification failed: invalid SovereignKey size %d", len(sk))
	}
	return nil
}

func (pe *DefaultPolicyEngine) AllowsLineage(lineage string, role string) error {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	allowed, ok := pe.AllowedLineages[role]
	if !ok {
		return fmt.Errorf("no lineage policy configured for role %s", role)
	}
	for _, ln := range allowed {
		if ln == lineage {
			return nil
		}
	}
	return fmt.Errorf("lineage %s is unauthorized for role %s", lineage, role)
}

func (pe *DefaultPolicyEngine) UpdateFromConfig(cfg PolicyConfig) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if cfg.LineageRules != nil {
		pe.AllowedLineages = cfg.LineageRules
	}
	return nil
}

// ==========================================
// 2. Policy File Loader, Snapshotting, and Hot-Reload Watcher
// ==========================================

func LoadPolicyConfig(path string) (PolicyConfig, error) {
	var cfg PolicyConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func SnapshotPolicy(cfg PolicyConfig, snapshotDir string) error {
	err := os.MkdirAll(snapshotDir, 0755)
	if err != nil {
		return err
	}
	fname := filepath.Join(snapshotDir, fmt.Sprintf("policy_%s.json", cfg.Version))
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fname, data, 0644)
}

func StartPolicyWatcher(ctx context.Context, path string, engine PolicyEngine, logger *log.Logger) {
	var lastMod time.Time

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastMod) {
				cfg, err := LoadPolicyConfig(path)
				if err != nil {
					logger.Printf("[PolicyWatcher] Load error from %s: %v\n", path, err)
					continue
				}

				if err := engine.UpdateFromConfig(cfg); err != nil {
					logger.Printf("[PolicyWatcher] Update error: %v\n", err)
					continue
				}

				lastMod = info.ModTime()
				logger.Printf("[PolicyWatcher] Policy rules (Version: %s) updated successfully from %s at %s\n", cfg.Version, path, lastMod.Format(time.Kitchen))
			}
		}
	}
}

// ==========================================
// 3. Swarm Supervisor Implementation
// ==========================================

type SwarmSupervisor interface {
	RegisterAgent(spec AgentSpec, id *pikr.Identity5) error
	StartSwarm(ctx context.Context) error
	StopSwarm() error
	GetAgentStatus(name string) (AgentStatus, int, error)
	SimulateAgentFailure(name string, err error)
	GetPolicyEngine() PolicyEngine
}

type swarmSupervisor struct {
	mu         sync.RWMutex
	agents     map[string]*AgentInstance
	logger     *log.Logger
	shutdownCh chan struct{}
	policy     PolicyEngine
}

func NewSwarmSupervisor() SwarmSupervisor {
	return &swarmSupervisor{
		agents:     make(map[string]*AgentInstance),
		logger:     log.New(os.Stdout, "[SwarmSupervisor] ", log.LstdFlags),
		shutdownCh: make(chan struct{}),
		policy:     NewDefaultPolicyEngine(),
	}
}

func NewSwarmSupervisorWithPolicy(pe PolicyEngine) SwarmSupervisor {
	return &swarmSupervisor{
		agents:     make(map[string]*AgentInstance),
		logger:     log.New(os.Stdout, "[SwarmSupervisor] ", log.LstdFlags),
		shutdownCh: make(chan struct{}),
		policy:     pe,
	}
}

func (s *swarmSupervisor) GetPolicyEngine() PolicyEngine {
	return s.policy
}

func (s *swarmSupervisor) RegisterAgent(spec AgentSpec, id *pikr.Identity5) error {
	// Policy checks: Enforce identity & lineage bounds
	if err := s.policy.AllowsIdentity(id, spec.Role); err != nil {
		return fmt.Errorf("supervisor registration blocked: %w", err)
	}

	if err := s.policy.AllowsLineage(spec.Lineage, spec.Role); err != nil {
		return fmt.Errorf("supervisor registration blocked: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.agents[spec.Name]; ok {
		return fmt.Errorf("agent %s already registered", spec.Name)
	}

	s.agents[spec.Name] = &AgentInstance{
		Spec:     spec,
		Status:   StatusIdle,
		Identity: id,
		LastSeen: time.Now(),
	}

	s.logger.Printf("Registered agent spec: %s | Role: %s | Lineage: %s (Policy Checks Passed)\n", spec.Name, spec.Role, spec.Lineage)
	return nil
}

func (s *swarmSupervisor) StartSwarm(ctx context.Context) error {
	s.logger.Println("Starting autonomous agent swarm...")

	s.mu.Lock()
	for _, inst := range s.agents {
		inst.Status = StatusRunning
		inst.LastSeen = time.Now()
		s.logger.Printf("Launched agent thread for: %s\n", inst.Spec.Name)
	}
	s.mu.Unlock()

	// Spin up supervisor monitoring loop
	go s.monitor(ctx)

	return nil
}

func (s *swarmSupervisor) StopSwarm() error {
	s.logger.Println("Stopping autonomous agent swarm...")
	close(s.shutdownCh)

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, inst := range s.agents {
		inst.Status = StatusIdle
	}

	return nil
}

func (s *swarmSupervisor) GetAgentStatus(name string) (AgentStatus, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inst, ok := s.agents[name]
	if !ok {
		return StatusIdle, 0, fmt.Errorf("agent %s not found", name)
	}

	return inst.Status, inst.Errors, nil
}

func (s *swarmSupervisor) SimulateAgentFailure(name string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inst, ok := s.agents[name]
	if !ok {
		return
	}

	inst.Status = StatusError
	inst.Errors++
	inst.LastSeen = time.Now()
	s.logger.Printf("[Failure] Agent %s crashed: %v | Total restarts: %d\n", name, err, inst.Errors)
}

func (s *swarmSupervisor) monitor(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.shutdownCh:
			return
		case <-ticker.C:
			s.mu.Lock()
			for name, inst := range s.agents {
				if inst.Status == StatusError {
					if inst.Errors < inst.Spec.MaxRetries {
						s.logger.Printf("[Self-Healing] Restarting agent %s (Attempt %d/%d)...\n", name, inst.Errors+1, inst.Spec.MaxRetries)
						inst.Status = StatusRunning
						inst.LastSeen = time.Now()
					} else {
						s.logger.Printf("[Alert] Agent %s reached max retries limit (%d). Leaving in ERROR state.\n", name, inst.Spec.MaxRetries)
					}
				}
			}
			s.mu.Unlock()
		}
	}
}
