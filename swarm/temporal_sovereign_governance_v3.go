package main

import (
	"context"
	"sync"
	"time"
)

// =========================
// Core Identifiers & Types
// =========================

type HarmonicGovernancePolicyID string
type GovernanceSessionID string
type GovernanceEventID string

// =========================
// Governance Scopes & Policies
// =========================

type GovernanceScopeKind string

const (
	GovernanceScopeAtlas         GovernanceScopeKind = "ATLAS"
	GovernanceScopeLayer         GovernanceScopeKind = "LAYER"
	GovernanceScopeConstellation GovernanceScopeKind = "CONSTELLATION"
	GovernanceScopeIdentity      GovernanceScopeKind = "IDENTITY"
)

type GovernanceScope struct {
	Kind            GovernanceScopeKind  `json:"kind"`
	AtlasID         *AtlasID             `json:"atlas_id"`
	LayerID         *string              `json:"layer_id"`
	ConstellationID *string              `json:"constellation_id"`
	IdentityID      *SovereignIdentityID `json:"identity_id"`
	Attributes      map[string]string    `json:"attributes"`
}

type HarmonicGovernancePolicy struct {
	ID        HarmonicGovernancePolicyID `json:"id"`
	Name      string                     `json:"name"`
	Scope     GovernanceScope            `json:"scope"`
	Priority  int                        `json:"priority"`
	Rules     []string                   `json:"rules"`
	Metadata  map[string]string          `json:"metadata"`
	CreatedAt time.Time                  `json:"created_at"`
}

// =========================
// Governance Decisions & Sessions
// =========================

type GovernanceDecisionKind string

const (
	GovernanceDecisionAllow  GovernanceDecisionKind = "ALLOW"
	GovernanceDecisionDeny   GovernanceDecisionKind = "DENY"
	GovernanceDecisionAdjust GovernanceDecisionKind = "ADJUST"
)

type GovernanceDecision struct {
	Kind       GovernanceDecisionKind      `json:"kind"`
	Reason     string                      `json:"reason"`
	PolicyID   *HarmonicGovernancePolicyID `json:"policy_id"`
	Attributes map[string]string           `json:"attributes"`
	Coordinate TemporalCoordinate          `json:"coordinate"`
	CreatedAt  time.Time                   `json:"created_at"`
}

type GovernanceSession struct {
	ID         GovernanceSessionID        `json:"id"`
	AtlasID    AtlasID                    `json:"atlas_id"`
	Policies   []HarmonicGovernancePolicy `json:"policies"`
	Decisions  []GovernanceDecision       `json:"decisions"`
	StartedAt  time.Time                  `json:"started_at"`
	FinishedAt *time.Time                 `json:"finished_at"`
	Coordinate TemporalCoordinate         `json:"coordinate"`
	Metadata   map[string]string          `json:"metadata"`
}

// =========================
// Governance Stability Metrics
// =========================

type GovernanceStabilityMetric struct {
	PolicyCoherence        float64  `json:"policy_coherence"`
	MeshCompliance         float64  `json:"mesh_compliance"`
	TemporalContinuity     float64  `json:"temporal_continuity"`
	RealityConsistency     float64  `json:"reality_consistency"`
	ConflictResolutionRate float64  `json:"conflict_resolution_rate"`
	Tags                   []string `json:"tags"`
}

type GovernanceConstraint struct {
	MinPolicyCoherence   float64  `json:"min_policy_coherence"`
	MinMeshCompliance    float64  `json:"min_mesh_compliance"`
	MaxConflictRate      float64  `json:"max_conflict_rate"`
	MaxRealityDivergence float64  `json:"max_reality_divergence"`
	PolicyTags           []string `json:"policy_tags"`
}

type GovernanceEngineConfig struct {
	GlobalConstraints   GovernanceConstraint                          `json:"global_constraints"`
	IdentityConstraints map[SovereignIdentityID]GovernanceConstraint `json:"identity_constraints"`
	GovernanceEvalInterval time.Duration                              `json:"governance_eval_interval"`
	RepairCheckInterval    time.Duration                              `json:"repair_check_interval"`
	Metadata            map[string]string                             `json:"metadata"`
}

type GovernanceEvalRequest struct {
	AtlasID             AtlasID                      `json:"atlas_id"`
	ActivePolicies      []HarmonicGovernancePolicyID `json:"active_policies"`
	OverrideConstraints *GovernanceConstraint        `json:"override_constraints"`
}

type GovernanceEvalResult struct {
	AtlasID            AtlasID                   `json:"atlas_id"`
	Metric             GovernanceStabilityMetric `json:"metric"`
	Session            GovernanceSession         `json:"session"`
	Decisions          []GovernanceDecision      `json:"decisions"`
	RecommendedActions []HarmonicGovernanceAction `json:"recommended_actions"`
}

type HarmonicGovernanceActionKind string

const (
	GovernanceActionAddPolicy    HarmonicGovernanceActionKind = "ADD_POLICY"
	GovernanceActionRemovePolicy HarmonicGovernanceActionKind = "REMOVE_POLICY"
	GovernanceActionRetunePolicy HarmonicGovernanceActionKind = "RETUNE_POLICY"
	GovernanceActionRebalance    HarmonicGovernanceActionKind = "REBALANCE"
)

type HarmonicGovernanceAction struct {
	Description string                       `json:"description"`
	AtlasID     *AtlasID                     `json:"atlas_id"`
	PolicyID    *HarmonicGovernancePolicyID  `json:"policy_id"`
	Kind        HarmonicGovernanceActionKind `json:"kind"`
	Params      map[string]any               `json:"params"`
	Priority    int                          `json:"priority"`
}

type HarmonicGovernancePlan struct {
	AtlasID    AtlasID                   `json:"atlas_id"`
	Actions    []HarmonicGovernanceAction `json:"actions"`
	Metric     GovernanceStabilityMetric `json:"metric"`
	Metadata   map[string]string         `json:"metadata"`
	CreatedAt  time.Time                 `json:"created_at"`
	Coordinate TemporalCoordinate        `json:"coordinate"`
	Version    string                    `json:"version"`
}

// =========================
// Storage & Registry Layer
// =========================

type GovernanceRegistry interface {
	StorePolicy(ctx context.Context, p HarmonicGovernancePolicy) error
	GetPolicy(ctx context.Context, id HarmonicGovernancePolicyID) (*HarmonicGovernancePolicy, error)
	ListPolicies(ctx context.Context, atlasID AtlasID) ([]HarmonicGovernancePolicy, error)
	StoreSession(ctx context.Context, s GovernanceSession) error
	GetSession(ctx context.Context, id GovernanceSessionID) (*GovernanceSession, error)
	ListSessions(ctx context.Context, atlasID AtlasID) ([]GovernanceSession, error)
}

type GovernanceEventType string

const (
	GovernanceEventSessionStarted  GovernanceEventType = "SESSION_STARTED"
	GovernanceEventSessionFinished GovernanceEventType = "SESSION_FINISHED"
	GovernanceEventPolicyChanged   GovernanceEventType = "POLICY_CHANGED"
	GovernanceEventRepaired        GovernanceEventType = "GOVERNANCE_REPAIRED"
)

type GovernanceEvent struct {
	EventID   GovernanceEventID   `json:"event_id"`
	SessionID *GovernanceSessionID `json:"session_id"`
	AtlasID   AtlasID             `json:"atlas_id"`
	Type      GovernanceEventType `json:"type"`
	Coordinate TemporalCoordinate  `json:"coordinate"`
	Payload   map[string]any      `json:"payload"`
}

type GovernanceEventLog interface {
	IngestEvent(ctx context.Context, event GovernanceEvent) error
	StreamEvents(ctx context.Context, atlasID AtlasID) (<-chan GovernanceEvent, error)
}

// =========================
// Analysis, Planning & Repair Interfaces
// =========================

type GovernanceAnalyzer interface {
	Analyze(ctx context.Context, req GovernanceEvalRequest, constraints GovernanceConstraint) (*GovernanceEvalResult, error)
}

type GovernancePlanner interface {
	BuildPlan(ctx context.Context, result *GovernanceEvalResult) (*HarmonicGovernancePlan, error)
	ExecutePlan(ctx context.Context, plan *HarmonicGovernancePlan) error
}

type GovernanceRepairEngine interface {
	BuildRepairPlan(ctx context.Context, result *GovernanceEvalResult) (*HarmonicGovernancePlan, error)
	ExecuteRepairPlan(ctx context.Context, plan *HarmonicGovernancePlan) error
}

// =========================
// Engine Interface
// =========================

type IdentityHarmonicGovernanceEngine interface {
	Init(ctx context.Context, cfg GovernanceEngineConfig) error
	EvaluateGovernance(ctx context.Context, req GovernanceEvalRequest) (*HarmonicGovernancePlan, error)
	RepairGovernance(ctx context.Context, id AtlasID) (*HarmonicGovernancePlan, error)
	GetLatestSession(ctx context.Context, id GovernanceSessionID) (*GovernanceSession, error)
	ListPolicies(ctx context.Context, atlasID AtlasID) ([]HarmonicGovernancePolicy, error)
	ExportState(ctx context.Context) (map[string]any, error)
}

// =========================
// Default Engine Implementation
// =========================

type DefaultIdentityHarmonicGovernanceEngine struct {
	cfg GovernanceEngineConfig

	registry GovernanceRegistry
	eventLog GovernanceEventLog

	analyzer GovernanceAnalyzer
	planner  GovernancePlanner
	repairer GovernanceRepairEngine
}

func NewDefaultIdentityHarmonicGovernanceEngine(
	gr GovernanceRegistry,
	el GovernanceEventLog,
	an GovernanceAnalyzer,
	pl GovernancePlanner,
	re GovernanceRepairEngine,
) *DefaultIdentityHarmonicGovernanceEngine {
	return &DefaultIdentityHarmonicGovernanceEngine{
		registry: gr,
		eventLog: el,
		analyzer: an,
		planner:  pl,
		repairer: re,
	}
}

func NewDefaultIdentityHarmonicGovernanceEngineWithMocks() *DefaultIdentityHarmonicGovernanceEngine {
	gr := NewInMemoryGovernanceRegistry()
	el := NewInMemoryGovernanceEventLog()
	an := NewNoopGovernanceAnalyzer()
	pl := NewNoopGovernancePlanner()
	re := NewNoopGovernanceRepairEngine()

	return NewDefaultIdentityHarmonicGovernanceEngine(gr, el, an, pl, re)
}

func (e *DefaultIdentityHarmonicGovernanceEngine) Init(ctx context.Context, cfg GovernanceEngineConfig) error {
	e.cfg = cfg
	return nil
}

func (e *DefaultIdentityHarmonicGovernanceEngine) EvaluateGovernance(ctx context.Context, req GovernanceEvalRequest) (*HarmonicGovernancePlan, error) {
	constraints := e.resolveConstraints(req.AtlasID, req.OverrideConstraints)

	result, err := e.analyzer.Analyze(ctx, req, constraints)
	if err != nil {
		return nil, err
	}

	plan, err := e.planner.BuildPlan(ctx, result)
	if err != nil {
		return nil, err
	}

	if err := e.planner.ExecutePlan(ctx, plan); err != nil {
		return nil, err
	}

	evt := GovernanceEvent{
		EventID:   GovernanceEventID("gov-eval-" + time.Now().Format(time.RFC3339Nano)),
		SessionID: &result.Session.ID,
		AtlasID:   req.AtlasID,
		Type:      GovernanceEventSessionFinished,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   time.Now(),
			Epoch:       "Phase89",
		},
		Payload: map[string]any{"plan_version": plan.Version},
	}
	_ = e.eventLog.IngestEvent(ctx, evt)

	return plan, nil
}

func (e *DefaultIdentityHarmonicGovernanceEngine) RepairGovernance(ctx context.Context, id AtlasID) (*HarmonicGovernancePlan, error) {
	req := GovernanceEvalRequest{AtlasID: id}
	constraints := e.resolveConstraints(id, nil)

	result, err := e.analyzer.Analyze(ctx, req, constraints)
	if err != nil {
		return nil, err
	}

	plan, err := e.repairer.BuildRepairPlan(ctx, result)
	if err != nil {
		return nil, err
	}

	if err := e.repairer.ExecuteRepairPlan(ctx, plan); err != nil {
		return nil, err
	}

	evt := GovernanceEvent{
		EventID:   GovernanceEventID("gov-repair-" + time.Now().Format(time.RFC3339Nano)),
		SessionID: nil,
		AtlasID:   id,
		Type:      GovernanceEventRepaired,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   time.Now(),
			Epoch:       "Phase89",
		},
		Payload: map[string]any{"plan_version": plan.Version},
	}
	_ = e.eventLog.IngestEvent(ctx, evt)

	return plan, nil
}

func (e *DefaultIdentityHarmonicGovernanceEngine) GetLatestSession(ctx context.Context, id GovernanceSessionID) (*GovernanceSession, error) {
	return e.registry.GetSession(ctx, id)
}

func (e *DefaultIdentityHarmonicGovernanceEngine) ListPolicies(ctx context.Context, atlasID AtlasID) ([]HarmonicGovernancePolicy, error) {
	return e.registry.ListPolicies(ctx, atlasID)
}

func (e *DefaultIdentityHarmonicGovernanceEngine) ExportState(ctx context.Context) (map[string]any, error) {
	return map[string]any{
		"config": e.cfg,
	}, nil
}

func (e *DefaultIdentityHarmonicGovernanceEngine) resolveConstraints(atlasID AtlasID, override *GovernanceConstraint) GovernanceConstraint {
	if override != nil {
		return *override
	}
	return e.cfg.GlobalConstraints
}

// =========================
// Mocks for Compilation
// =========================

type InMemoryGovernanceRegistry struct {
	mu          sync.RWMutex
	policies    map[HarmonicGovernancePolicyID]HarmonicGovernancePolicy
	sessions    map[GovernanceSessionID]GovernanceSession
	sessByAtlas map[AtlasID][]GovernanceSession
}

func NewInMemoryGovernanceRegistry() *InMemoryGovernanceRegistry {
	return &InMemoryGovernanceRegistry{
		policies:    make(map[HarmonicGovernancePolicyID]HarmonicGovernancePolicy),
		sessions:    make(map[GovernanceSessionID]GovernanceSession),
		sessByAtlas: make(map[AtlasID][]GovernanceSession),
	}
}

func (r *InMemoryGovernanceRegistry) StorePolicy(ctx context.Context, p HarmonicGovernancePolicy) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.policies[p.ID] = p
	return nil
}

func (r *InMemoryGovernanceRegistry) GetPolicy(ctx context.Context, id HarmonicGovernancePolicyID) (*HarmonicGovernancePolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.policies[id]
	if !ok {
		return nil, nil
	}
	return &p, nil
}

func (r *InMemoryGovernanceRegistry) ListPolicies(ctx context.Context, atlasID AtlasID) ([]HarmonicGovernancePolicy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []HarmonicGovernancePolicy
	for _, p := range r.policies {
		if p.Scope.AtlasID != nil && *p.Scope.AtlasID == atlasID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (r *InMemoryGovernanceRegistry) StoreSession(ctx context.Context, s GovernanceSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[s.ID] = s
	r.sessByAtlas[s.AtlasID] = append(r.sessByAtlas[s.AtlasID], s)
	return nil
}

func (r *InMemoryGovernanceRegistry) GetSession(ctx context.Context, id GovernanceSessionID) (*GovernanceSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[id]
	if !ok {
		return nil, nil
	}
	return &s, nil
}

func (r *InMemoryGovernanceRegistry) ListSessions(ctx context.Context, atlasID AtlasID) ([]GovernanceSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]GovernanceSession(nil), r.sessByAtlas[atlasID]...), nil
}

type InMemoryGovernanceEventLog struct {
	mu     sync.RWMutex
	events []GovernanceEvent
}

func NewInMemoryGovernanceEventLog() *InMemoryGovernanceEventLog {
	return &InMemoryGovernanceEventLog{
		events: make([]GovernanceEvent, 0),
	}
}

func (l *InMemoryGovernanceEventLog) IngestEvent(ctx context.Context, event GovernanceEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
	return nil
}

func (l *InMemoryGovernanceEventLog) StreamEvents(ctx context.Context, atlasID AtlasID) (<-chan GovernanceEvent, error) {
	out := make(chan GovernanceEvent)
	go func() {
		defer close(out)
		l.mu.RLock()
		defer l.mu.RUnlock()
		for _, e := range l.events {
			if e.AtlasID == atlasID {
				select {
				case <-ctx.Done():
					return
				case out <- e:
				}
			}
		}
	}()
	return out, nil
}

type NoopGovernanceAnalyzer struct{}

func NewNoopGovernanceAnalyzer() *NoopGovernanceAnalyzer {
	return &NoopGovernanceAnalyzer{}
}

func (a *NoopGovernanceAnalyzer) Analyze(
	ctx context.Context,
	req GovernanceEvalRequest,
	constraints GovernanceConstraint,
) (*GovernanceEvalResult, error) {
	sessionID := GovernanceSessionID("session-" + time.Now().Format(time.RFC3339Nano))
	now := time.Now()
	coord := TemporalCoordinate{LogicalTick: 0, WallClock: now, Epoch: "Phase89"}

	session := GovernanceSession{
		ID:         sessionID,
		AtlasID:    req.AtlasID,
		Policies:   nil,
		Decisions:  nil,
		StartedAt:  now,
		FinishedAt: &now,
		Coordinate: coord,
		Metadata:   map[string]string{"analyzer": "noop"},
	}

	metric := GovernanceStabilityMetric{
		PolicyCoherence:        1.0,
		MeshCompliance:         1.0,
		TemporalContinuity:     1.0,
		RealityConsistency:     1.0,
		ConflictResolutionRate: 1.0,
		Tags:                   []string{"noop", "stable"},
	}

	return &GovernanceEvalResult{
		AtlasID:            req.AtlasID,
		Metric:             metric,
		Session:            session,
		Decisions:          nil,
		RecommendedActions: nil,
	}, nil
}

type NoopGovernancePlanner struct{}

func NewNoopGovernancePlanner() *NoopGovernancePlanner {
	return &NoopGovernancePlanner{}
}

func (p *NoopGovernancePlanner) BuildPlan(ctx context.Context, result *GovernanceEvalResult) (*HarmonicGovernancePlan, error) {
	now := time.Now()
	plan := &HarmonicGovernancePlan{
		AtlasID:   result.AtlasID,
		Actions:   nil,
		Metric:    result.Metric,
		Metadata:  map[string]string{"planner": "noop"},
		CreatedAt: now,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   now,
			Epoch:       "Phase89",
		},
		Version: "0.0.1-noop",
	}
	return plan, nil
}

func (p *NoopGovernancePlanner) ExecutePlan(ctx context.Context, plan *HarmonicGovernancePlan) error {
	return nil
}

type NoopGovernanceRepairEngine struct{}

func NewNoopGovernanceRepairEngine() *NoopGovernanceRepairEngine {
	return &NoopGovernanceRepairEngine{}
}

func (r *NoopGovernanceRepairEngine) BuildRepairPlan(ctx context.Context, result *GovernanceEvalResult) (*HarmonicGovernancePlan, error) {
	now := time.Now()
	plan := &HarmonicGovernancePlan{
		AtlasID:   result.AtlasID,
		Actions:   nil,
		Metric:    result.Metric,
		Metadata:  map[string]string{"repairer": "noop"},
		CreatedAt: now,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   now,
			Epoch:       "Phase89",
		},
		Version: "0.0.1-noop",
	}
	return plan, nil
}

func (r *NoopGovernanceRepairEngine) ExecuteRepairPlan(ctx context.Context, plan *HarmonicGovernancePlan) error {
	return nil
}
