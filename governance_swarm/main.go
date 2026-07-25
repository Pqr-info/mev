package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	mathrand "math/rand"
	"sync"
	"time"

	"pqr.info/mev/governance"
	"pqr.info/jetweb-core/core/go/pikr"
)

// ==========================================
// 1. Swarm Daemon Interfaces (Copilot Blueprint)
// ==========================================

type Event struct {
	Type  string
	Value float64
}

type Proposal struct {
	TargetField string
	NewValue    interface{}
}

type Telemetry struct {
	Metric string
	Value  float64
}

type PolicyMutation struct {
	ID        string
	Author    *pikr.Identity5
	Role      string
	Target    string // e.g. "Sentinel"
	Operation string // "add", "remove"
	Value     string // e.g. "corridor.new_horizon"
	Reason    string
	Signature []byte
}

type GovernanceDaemonClient interface {
	ConnectAgent(ctx context.Context, id *pikr.Identity5, sk []byte, role string) (AgentSession, error)
}

type AgentSession interface {
	SubscribeEvents(ctx context.Context) (<-chan Event, error)
	SubmitProposal(ctx context.Context, p Proposal) error
	ReportTelemetry(ctx context.Context, t Telemetry) error
	SubmitPolicyMutation(ctx context.Context, pm PolicyMutation) error
}

// ==========================================
// 2. Swarm Daemon Core Server (Daemon + Client)
// ==========================================

type SwarmDaemon struct {
	mu            sync.RWMutex
	ce            *governance.ConstitutionEngine
	je            *governance.TemporalJudiciaryEngine
	sae           *governance.SanctionsAndAppealsEngine
	rce           *governance.RightsComplianceEngine
	supervisor    governance.SwarmSupervisor
	shutdownCh    chan struct{}
	activeAgents  map[string]*AgentConfig
	eventChs      map[string]chan Event
	mutationQueue []PolicyMutation
}

type AgentConfig struct {
	Name         string
	Role         string
	Identity     *pikr.Identity5
	SovereignKey string
}

func NewSwarmDaemon() *SwarmDaemon {
	constObj := governance.TemporalConstitution{
		ReplayLimits: governance.ReplayLimits{
			MaxReplayPerEpoch: 10,
			AllowFractured:    false,
		},
		TimeslipRules: governance.TimeslipRules{
			AllowTimeslips: false,
		},
		MutationBoundaries: governance.MutationBoundaries{
			MaxMutationRate: 0.12,
		},
		FederationObligations: governance.FederationObligations{
			MinSyncFrequencySeconds: 15,
		},
	}

	ce := governance.NewConstitutionEngine(constObj)
	je := governance.NewTemporalJudiciaryEngine(ce)
	sae := governance.NewSanctionsAndAppealsEngine()
	rce := governance.NewRightsComplianceEngine()
	sup := governance.NewSwarmSupervisor()

	return &SwarmDaemon{
		ce:            ce,
		je:            je,
		sae:           sae,
		rce:           rce,
		supervisor:    sup,
		shutdownCh:    make(chan struct{}),
		activeAgents:  make(map[string]*AgentConfig),
		eventChs:      make(map[string]chan Event),
		mutationQueue: make([]PolicyMutation, 0),
	}
}

// ConnectAgent handles identity and lineage verification (5D-PIKR invariance check)
func (sd *SwarmDaemon) ConnectAgent(ctx context.Context, id *pikr.Identity5, sk []byte, role string) (AgentSession, error) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	skStr := hex.EncodeToString(sk)
	agent, ok := sd.activeAgents[skStr]
	if !ok {
		return nil, errors.New("identity not pre-registered in daemon specs")
	}

	// 5D-PIKR Invariance check
	derivedSK := id.RecoveryMatrix().SovereignKey()
	if hex.EncodeToString(derivedSK) != skStr {
		return nil, errors.New("cryptographic signature verification failure: invalid SovereignKey")
	}

	// Verify rights & citizenship
	if !sd.rce.AssertRights(agent.Name) {
		return nil, errors.New("agent lacks active citizenship due-process authorization")
	}

	ch := make(chan Event, 10)
	sd.eventChs[agent.Name] = ch

	log.Printf("[Daemon] Cryptographic connection established for SK: %s (%s)\n", skStr[:12], agent.Name)
	return &sessionInstance{
		daemon:    sd,
		agentName: agent.Name,
		sk:        skStr,
		eventCh:   ch,
	}, nil
}

// ==========================================
// 3. Agent Session Instance (Copilot Session)
// ==========================================

type sessionInstance struct {
	daemon    *SwarmDaemon
	agentName string
	sk        string
	eventCh   chan Event
}

func (s *sessionInstance) SubscribeEvents(ctx context.Context) (<-chan Event, error) {
	return s.eventCh, nil
}

func (s *sessionInstance) SubmitProposal(ctx context.Context, p Proposal) error {
	s.daemon.mu.RLock()
	agent := s.daemon.activeAgents[s.sk]
	s.daemon.mu.RUnlock()

	// Only Analyst and Actuators can submit proposals
	if agent.Role != "Analyst" {
		s.daemon.sae.EnforceSanction(&governance.Sanction{
			SanctionID: fmt.Sprintf("sanc-%d", time.Now().UnixNano()),
			Actor:      agent.Name,
			Severity:   6,
			CreatedAt:  time.Now(),
		})
		log.Printf("[Daemon] [Denial] Sentinel agent %s attempted mutation. Applying Sanction.\n", agent.Name)
		return errors.New("unauthorized role mutation permission")
	}

	log.Printf("[Daemon] [Proposal] Ingesting MaxMutationRate change (%.4f) from Analyst %s...\n", p.NewValue, agent.Name)
	leg := governance.NewTemporalLegislature(s.daemon.ce)
	err := leg.ApplyConstitutionAmendment(governance.Amendment{
		Field:    p.TargetField,
		NewValue: p.NewValue,
	})
	if err != nil {
		log.Printf("[Daemon] [Failure] Legislature rejected amendment: %v\n", err)
	} else {
		log.Printf("[Daemon] [Success] Constitution successfully amended! New MaxMutationRate: %.4f\n", s.daemon.ce.GetConstitution().MutationBoundaries.MaxMutationRate)
	}
	return nil
}

func (s *sessionInstance) ReportTelemetry(ctx context.Context, t Telemetry) error {
	s.daemon.mu.RLock()
	agent := s.daemon.activeAgents[s.sk]
	s.daemon.mu.RUnlock()

	if t.Metric == "mutation_rate" && t.Value > s.daemon.ce.GetConstitution().MutationBoundaries.MaxMutationRate {
		log.Printf("[Daemon] [Anomaly] Telemetry anomaly: %.4f from agent %s. Adjudicating case...\n", t.Value, agent.Name)

		d := governance.TemporalGovernanceDirective{
			Reason:             "unstable-epoch",
			ReplayAllowed:      true,
			TimeslipAllowed:    false,
			MutationRateCap:    t.Value,
			FederationSyncFreq: 10 * time.Second,
		}

		ruling := s.daemon.je.Adjudicate(fmt.Sprintf("epoch-%d", time.Now().Unix()), agent.Name, d)
		log.Printf("[Judiciary] ruling ID %s Verdict: %s | Cap Corrected to: %.4f\n", ruling.RulingID, ruling.Verdict, ruling.Directive.MutationRateCap)

		// Violations prompt a corrective sanction
		s.daemon.sae.EnforceSanction(&governance.Sanction{
			SanctionID: fmt.Sprintf("sanc-%s", ruling.RulingID),
			Actor:      ruling.NodeID,
			Severity:   7,
			CreatedAt:  time.Now(),
		})
	}
	return nil
}

func (s *sessionInstance) SubmitPolicyMutation(ctx context.Context, pm PolicyMutation) error {
	s.daemon.mu.Lock()
	defer s.daemon.mu.Unlock()

	// Validate agent cryptographic authority (SovereignKey length validation)
	sk := pm.Author.RecoveryMatrix().SovereignKey()
	if len(sk) != 32 {
		return errors.New("cryptographic signature validation failed: invalid SovereignKey size")
	}

	// Enforce that only Analyst agents can submit policy mutation proposals
	if pm.Role != "Analyst" {
		return errors.New("unauthorized role: only Analyst agents may propose policy changes")
	}

	log.Printf("[Daemon] [PolicyProposal] Ingesting policy mutation from Analyst %s to %s lineage rules for %s (%s)...\n", s.agentName, pm.Operation, pm.Target, pm.Value)
	s.daemon.mutationQueue = append(s.daemon.mutationQueue, pm)
	return nil
}

// ==========================================
// 4. Register Swarm Specifications
// ==========================================

func (sd *SwarmDaemon) RegisterAgentSpec(name string, role string, lineage string) (*pikr.Identity5, []byte, error) {
	sd.mu.Lock()
	defer sd.mu.Unlock()

	psi := make([]byte, 32)
	sfi := make([]byte, 16)
	tsi := make([]byte, 16)
	qii := make([]byte, 32)
	qri := make([]byte, 8)
	rand.Read(psi)
	rand.Read(sfi)
	rand.Read(tsi)
	rand.Read(qii)
	rand.Read(qri)

	id, err := pikr.NewIdentity5(psi, sfi, tsi, qii, qri)
	if err != nil {
		return nil, nil, err
	}

	sk := id.RecoveryMatrix().SovereignKey()
	skStr := hex.EncodeToString(sk)

	sd.activeAgents[skStr] = &AgentConfig{
		Name:         name,
		Role:         role,
		Identity:     id,
		SovereignKey: skStr,
	}

	// Active citizenship
	sd.rce.RegisterCitizen(&governance.Citizenship{
		NodeID:      name,
		Authorized:  true,
		EpochJoined: 1,
	})

	// Register in Swarm Supervisor
	sd.supervisor.RegisterAgent(governance.AgentSpec{
		Name:       name,
		Role:       role,
		Lineage:    lineage,
		MaxRetries: 3,
	}, id)

	return id, sk, nil
}

// StartMutationEvaluator evaluates queued policy proposals asynchronously
func (sd *SwarmDaemon) StartMutationEvaluator(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(800 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sd.mu.Lock()
				if len(sd.mutationQueue) > 0 {
					pm := sd.mutationQueue[0]
					sd.mutationQueue = sd.mutationQueue[1:]

					log.Printf("[Daemon] [Evaluator] Evaluating policy mutation: Approve %s to %s rules (%s)\n", pm.Value, pm.Target, pm.Reason)

					// Approved! Mutate active policy rules
					engine := sd.supervisor.GetPolicyEngine()
					cfg := governance.PolicyConfig{
						LineageRules: map[string][]string{
							pm.Target: {pm.Value},
						},
					}
					engine.UpdateFromConfig(cfg)

					log.Printf("[Daemon] [Evaluator] Policy Mutation Approved and Applied! Swarm Policy Engine updated.\n")
				}
				sd.mu.Unlock()
			}
		}
	}()
}

// ==========================================
// 5. Simulation Agent Runtime (Wired to Connect)
// ==========================================

func RunAgentProcess(ctx context.Context, daemon *SwarmDaemon, name string, id *pikr.Identity5, sk []byte, role string) {
	session, err := daemon.ConnectAgent(ctx, id, sk, role)
	if err != nil {
		log.Printf("[Agent %s] [Error] Connection failed: %v\n", name, err)
		return
	}

	events, _ := session.SubscribeEvents(ctx)

	go func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		defer ticker.Stop()

		proposalSubmitted := false

		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-events:
				log.Printf("[Agent %s] Received broadcast event: %s | Val: %.4f\n", name, ev.Type, ev.Value)
			case <-ticker.C:
				if role == "Sentinel" {
					// Emits telemetry
					val := 0.05 + mathrand.Float64()*0.15
					session.ReportTelemetry(ctx, Telemetry{Metric: "mutation_rate", Value: val})
				} else if role == "Analyst" {
					if !proposalSubmitted {
						// Propose a dynamic lineage expansion policy change via SubmitPolicyMutation
						err := session.SubmitPolicyMutation(ctx, PolicyMutation{
							ID:        fmt.Sprintf("mut-%d", time.Now().UnixNano()),
							Author:    id,
							Role:      role,
							Target:    "Sentinel",
							Operation: "add",
							Value:     "corridor.new_horizon",
							Reason:    "expand swarm corridor bounds to new horizon",
						})
						if err == nil {
							proposalSubmitted = true
						}
					}

					// Proposes constitution amendments
					newCap := 0.15 + mathrand.Float64()*0.15
					session.SubmitProposal(ctx, Proposal{TargetField: "MaxMutationRate", NewValue: newCap})
				}
			}
		}
	}()
}

// ==========================================
// 6. Simulation Entry point
// ==========================================

func main() {
	mathrand.Seed(time.Now().UnixNano())
	log.Println("[SwarmDaemon] Booting up sovereign organism swarm controller...")

	daemon := NewSwarmDaemon()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Register specs in daemon and supervisor
	idAlpha, skAlpha, _ := daemon.RegisterAgentSpec("SentinelNode-Alpha", "Sentinel", "corridor.east")
	idBeta, skBeta, _ := daemon.RegisterAgentSpec("AnalystNode-Beta", "Analyst", "corridor.west")

	// Start swarm supervisor
	daemon.supervisor.StartSwarm(ctx)

	// Start policy mutation evaluator
	daemon.StartMutationEvaluator(ctx)

	// Launch autonomous process loops using ConnectAgent session bindings
	go RunAgentProcess(ctx, daemon, "SentinelNode-Alpha", idAlpha, skAlpha, "Sentinel")
	go RunAgentProcess(ctx, daemon, "AnalystNode-Beta", idBeta, skBeta, "Analyst")

	// Periodically broadcast events from daemon to sessions
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				daemon.mu.RLock()
				for name, ch := range daemon.eventChs {
					ch <- Event{Type: "epoch_tick", Value: float64(time.Now().Unix())}
					_ = name
				}
				daemon.mu.RUnlock()
			}
		}
	}()

	// Run simulation for 5 seconds
	time.Sleep(5 * time.Second)
	daemon.supervisor.StopSwarm()
	log.Println("[SwarmDaemon] Autonomous swarm shutdown completed cleanly.")
}
