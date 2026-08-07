package main

import (
	"context"
	"sync"
	"time"
)

// =========================
// Core Identifiers & Types
// =========================

type IdentityCrystalID string
type CrystalPhase string
type CrystalEventID string

const (
	CrystalPhaseLiquid   CrystalPhase = "LIQUID"
	CrystalPhaseSolid    CrystalPhase = "SOLID"
	CrystalPhaseAnnealed CrystalPhase = "ANNEALED"
)

type IdentityCrystal struct {
	ID         IdentityCrystalID   `json:"id"`
	Identity   SovereignIdentityID `json:"identity"`
	Phase      CrystalPhase        `json:"phase"`
	Stability  float64             `json:"stability"`
	Invariants map[string]any      `json:"invariants"`
	Attributes map[string]string   `json:"attributes"`
	CreatedAt  time.Time           `json:"created_at"`
	Coordinate TemporalCoordinate  `json:"coordinate"`
}

type IdentityCrystalSnapshot struct {
	CrystalID  IdentityCrystalID   `json:"crystal_id"`
	Identity   SovereignIdentityID `json:"identity"`
	Coordinate TemporalCoordinate  `json:"coordinate"`
	Phase      CrystalPhase        `json:"phase"`
	Stability  float64             `json:"stability"`
	Metadata   map[string]string   `json:"metadata"`
}

// =========================
// Metrics & Constraints
// =========================

type CrystallizationMetric struct {
	QualityScore     float64  `json:"quality_score"`
	PreStability     float64  `json:"pre_stability"`
	PostStability    float64  `json:"post_stability"`
	CompressionRatio float64  `json:"compression_ratio"`
	Tags             []string `json:"tags"`
}

type CrystallizationConstraint struct {
	MinStabilityForCrystallization float64  `json:"min_stability_for_crystallization"`
	MaxInvariantLoss               float64  `json:"max_invariant_loss"`
	PolicyTags                     []string `json:"policy_tags"`
}

// =========================
// Engine Configuration
// =========================

type CrystallizationEngineConfig struct {
	GlobalConstraints            CrystallizationConstraint                          `json:"global_constraints"`
	IdentityConstraints          map[SovereignIdentityID]CrystallizationConstraint `json:"identity_constraints"`
	CrystallizationCheckInterval time.Duration                                      `json:"crystallization_check_interval"`
	AnnealingCheckInterval       time.Duration                                      `json:"annealing_check_interval"`
	Metadata                     map[string]string                                  `json:"metadata"`
}

// =========================
// Requests & Results
// =========================

type CrystallizationRequest struct {
	Identity            SovereignIdentityID        `json:"identity"`
	OverrideConstraints *CrystallizationConstraint `json:"override_constraints"`
}

type CrystallizationResult struct {
	Identity           SovereignIdentityID     `json:"identity"`
	Metric             CrystallizationMetric   `json:"metric"`
	Crystal            IdentityCrystal         `json:"crystal"`
	Snapshot           IdentityCrystalSnapshot `json:"snapshot"`
	RecommendedActions []string                `json:"recommended_actions"`
}

type CrystalPlan struct {
	Identity   SovereignIdentityID   `json:"identity"`
	Actions    []string              `json:"actions"`
	Metric     CrystallizationMetric `json:"metric"`
	Metadata   map[string]string     `json:"metadata"`
	CreatedAt  time.Time             `json:"created_at"`
	Coordinate TemporalCoordinate    `json:"coordinate"`
	Version    string                `json:"version"`
}

// =========================
// Storage & Registry Layer
// =========================

type CrystalRegistry interface {
	StoreCrystal(ctx context.Context, crystal IdentityCrystal) error
	GetCrystal(ctx context.Context, id IdentityCrystalID) (*IdentityCrystal, error)
	GetLatestCrystal(ctx context.Context, identity SovereignIdentityID) (*IdentityCrystal, error)
	StoreSnapshot(ctx context.Context, snapshot IdentityCrystalSnapshot) error
	ListSnapshots(ctx context.Context, crystalID IdentityCrystalID) ([]IdentityCrystalSnapshot, error)
}

// =========================
// Crystallization Events
// =========================

type CrystallizationEventType string

const (
	CrystallizationEventCreated        CrystallizationEventType = "CRYSTAL_CREATED"
	CrystallizationEventAnnealed       CrystallizationEventType = "CRYSTAL_ANNEALED"
	CrystallizationEventDecrystallized CrystallizationEventType = "CRYSTAL_DECRYSTALLIZED"
)

type CrystallizationEvent struct {
	EventID    CrystalEventID           `json:"event_id"`
	Identity   SovereignIdentityID      `json:"identity"`
	CrystalID  *IdentityCrystalID       `json:"crystal_id"`
	Type       CrystallizationEventType `json:"type"`
	Coordinate TemporalCoordinate       `json:"coordinate"`
	Payload    map[string]any           `json:"payload"`
}

type CrystallizationEventLog interface {
	IngestEvent(ctx context.Context, event CrystallizationEvent) error
	StreamEvents(ctx context.Context, identity SovereignIdentityID) (<-chan CrystallizationEvent, error)
}

// =========================
// Analysis, Planning & Execution
// =========================

type CrystallizationAnalyzer interface {
	Analyze(ctx context.Context, req CrystallizationRequest, constraints CrystallizationConstraint) (*CrystallizationResult, error)
}

type CrystallizationPlanner interface {
	BuildPlan(ctx context.Context, result *CrystallizationResult) (*CrystalPlan, error)
}

type CrystallizationExecutor interface {
	ExecutePlan(ctx context.Context, plan *CrystalPlan) error
}

// =========================
// Engine Interface
// =========================

type IdentityCrystallizationEngine interface {
	Init(ctx context.Context, cfg CrystallizationEngineConfig) error
	CrystallizeIdentity(ctx context.Context, identity SovereignIdentityID) (*CrystalPlan, error)
	AnnealIdentity(ctx context.Context, identity SovereignIdentityID) (*CrystalPlan, error)
	DecrystallizeIdentity(ctx context.Context, identity SovereignIdentityID) (*CrystalPlan, error)
	GetLatestCrystal(ctx context.Context, identity SovereignIdentityID) (*IdentityCrystal, error)
	CrystallizeAll(ctx context.Context) error
	AnnealAll(ctx context.Context) error
	ExportState(ctx context.Context) (map[string]any, error)
}

// =========================
// Default Engine Implementation
// =========================

type DefaultIdentityCrystallizationEngine struct {
	cfg CrystallizationEngineConfig

	crystalRegistry CrystalRegistry
	eventLog        CrystallizationEventLog
	analyzer        CrystallizationAnalyzer
	planner         CrystallizationPlanner
	executor        CrystallizationExecutor
}

func NewDefaultIdentityCrystallizationEngine(
	cr CrystalRegistry,
	el CrystallizationEventLog,
	an CrystallizationAnalyzer,
	pl CrystallizationPlanner,
	ex CrystallizationExecutor,
) *DefaultIdentityCrystallizationEngine {
	return &DefaultIdentityCrystallizationEngine{
		crystalRegistry: cr,
		eventLog:        el,
		analyzer:        an,
		planner:         pl,
		executor:        ex,
	}
}

func (e *DefaultIdentityCrystallizationEngine) Init(ctx context.Context, cfg CrystallizationEngineConfig) error {
	e.cfg = cfg
	return nil
}

func (e *DefaultIdentityCrystallizationEngine) CrystallizeIdentity(ctx context.Context, identity SovereignIdentityID) (*CrystalPlan, error) {
	req := CrystallizationRequest{Identity: identity}
	constraints := e.resolveConstraints(identity, nil)

	result, err := e.analyzer.Analyze(ctx, req, constraints)
	if err != nil {
		return nil, err
	}

	plan, err := e.planner.BuildPlan(ctx, result)
	if err != nil {
		return nil, err
	}

	if err := e.executor.ExecutePlan(ctx, plan); err != nil {
		return nil, err
	}

	evt := CrystallizationEvent{
		EventID:   CrystalEventID("crystallize-" + string(identity)),
		Identity:  identity,
		CrystalID: &result.Crystal.ID,
		Type:      CrystallizationEventCreated,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   time.Now(),
			Epoch:       "Phase84",
		},
		Payload: map[string]any{"plan_version": plan.Version},
	}
	_ = e.eventLog.IngestEvent(ctx, evt)

	return plan, nil
}

func (e *DefaultIdentityCrystallizationEngine) AnnealIdentity(ctx context.Context, identity SovereignIdentityID) (*CrystalPlan, error) {
	req := CrystallizationRequest{Identity: identity}
	constraints := e.resolveConstraints(identity, nil)

	result, err := e.analyzer.Analyze(ctx, req, constraints)
	if err != nil {
		return nil, err
	}

	plan, err := e.planner.BuildPlan(ctx, result)
	if err != nil {
		return nil, err
	}

	if err := e.executor.ExecutePlan(ctx, plan); err != nil {
		return nil, err
	}

	evt := CrystallizationEvent{
		EventID:   CrystalEventID("anneal-" + string(identity)),
		Identity:  identity,
		CrystalID: nil,
		Type:      CrystallizationEventAnnealed,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   time.Now(),
			Epoch:       "Phase84",
		},
		Payload: map[string]any{"plan_version": plan.Version},
	}
	_ = e.eventLog.IngestEvent(ctx, evt)

	return plan, nil
}

func (e *DefaultIdentityCrystallizationEngine) DecrystallizeIdentity(ctx context.Context, identity SovereignIdentityID) (*CrystalPlan, error) {
	req := CrystallizationRequest{Identity: identity}
	constraints := e.resolveConstraints(identity, nil)

	result, err := e.analyzer.Analyze(ctx, req, constraints)
	if err != nil {
		return nil, err
	}

	plan, err := e.planner.BuildPlan(ctx, result)
	if err != nil {
		return nil, err
	}

	if err := e.executor.ExecutePlan(ctx, plan); err != nil {
		return nil, err
	}

	evt := CrystallizationEvent{
		EventID:   CrystalEventID("decrystallize-" + string(identity)),
		Identity:  identity,
		CrystalID: nil,
		Type:      CrystallizationEventDecrystallized,
		Coordinate: TemporalCoordinate{
			LogicalTick: 0,
			WallClock:   time.Now(),
			Epoch:       "Phase84",
		},
		Payload: map[string]any{"plan_version": plan.Version},
	}
	_ = e.eventLog.IngestEvent(ctx, evt)

	return plan, nil
}

func (e *DefaultIdentityCrystallizationEngine) GetLatestCrystal(ctx context.Context, identity SovereignIdentityID) (*IdentityCrystal, error) {
	return e.crystalRegistry.GetLatestCrystal(ctx, identity)
}

func (e *DefaultIdentityCrystallizationEngine) CrystallizeAll(ctx context.Context) error {
	return nil
}

func (e *DefaultIdentityCrystallizationEngine) AnnealAll(ctx context.Context) error {
	return nil
}

func (e *DefaultIdentityCrystallizationEngine) ExportState(ctx context.Context) (map[string]any, error) {
	return map[string]any{
		"config": e.cfg,
	}, nil
}

func (e *DefaultIdentityCrystallizationEngine) resolveConstraints(id SovereignIdentityID, override *CrystallizationConstraint) CrystallizationConstraint {
	if override != nil {
		return *override
	}
	if e.cfg.IdentityConstraints != nil {
		if c, ok := e.cfg.IdentityConstraints[id]; ok {
			return c
		}
	}
	return e.cfg.GlobalConstraints
}

// =========================
// Mocks for Compilation & Testing
// =========================

type InMemoryCrystalRegistry struct {
	mu         sync.RWMutex
	crystals   map[IdentityCrystalID]IdentityCrystal
	byIdentity map[SovereignIdentityID]IdentityCrystalID
	snapshots  map[IdentityCrystalID][]IdentityCrystalSnapshot
}

func NewInMemoryCrystalRegistry() *InMemoryCrystalRegistry {
	return &InMemoryCrystalRegistry{
		crystals:   make(map[IdentityCrystalID]IdentityCrystal),
		byIdentity: make(map[SovereignIdentityID]IdentityCrystalID),
		snapshots:  make(map[IdentityCrystalID][]IdentityCrystalSnapshot),
	}
}

func (r *InMemoryCrystalRegistry) StoreCrystal(ctx context.Context, crystal IdentityCrystal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.crystals[crystal.ID] = crystal
	r.byIdentity[crystal.Identity] = crystal.ID
	return nil
}

func (r *InMemoryCrystalRegistry) GetCrystal(ctx context.Context, id IdentityCrystalID) (*IdentityCrystal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.crystals[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

func (r *InMemoryCrystalRegistry) GetLatestCrystal(ctx context.Context, identity SovereignIdentityID) (*IdentityCrystal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byIdentity[identity]
	if !ok {
		return nil, nil
	}
	c, ok := r.crystals[id]
	if !ok {
		return nil, nil
	}
	return &c, nil
}

func (r *InMemoryCrystalRegistry) StoreSnapshot(ctx context.Context, snapshot IdentityCrystalSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshots[snapshot.CrystalID] = append(r.snapshots[snapshot.CrystalID], snapshot)
	return nil
}

func (r *InMemoryCrystalRegistry) ListSnapshots(ctx context.Context, crystalID IdentityCrystalID) ([]IdentityCrystalSnapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]IdentityCrystalSnapshot(nil), r.snapshots[crystalID]...), nil
}

type InMemoryCrystallizationEventLog struct {
	mu     sync.RWMutex
	events []CrystallizationEvent
}

func NewInMemoryCrystallizationEventLog() *InMemoryCrystallizationEventLog {
	return &InMemoryCrystallizationEventLog{
		events: make([]CrystallizationEvent, 0),
	}
}

func (l *InMemoryCrystallizationEventLog) IngestEvent(ctx context.Context, event CrystallizationEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = append(l.events, event)
	return nil
}

func (l *InMemoryCrystallizationEventLog) StreamEvents(ctx context.Context, identity SovereignIdentityID) (<-chan CrystallizationEvent, error) {
	out := make(chan CrystallizationEvent)
	go func() {
		defer close(out)
		l.mu.RLock()
		defer l.mu.RUnlock()
		for _, e := range l.events {
			if e.Identity == identity {
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

type NoopCrystallizationAnalyzer struct{}

func NewNoopCrystallizationAnalyzer() *NoopCrystallizationAnalyzer {
	return &NoopCrystallizationAnalyzer{}
}

func (a *NoopCrystallizationAnalyzer) Analyze(
	ctx context.Context,
	req CrystallizationRequest,
	constraints CrystallizationConstraint,
) (*CrystallizationResult, error) {
	crystal := IdentityCrystal{
		ID:         IdentityCrystalID("crystal-" + string(req.Identity)),
		Identity:   req.Identity,
		Phase:      CrystalPhaseSolid,
		Stability:  1.0,
		Invariants: map[string]any{"noop": true},
		Attributes: map[string]string{"analyzer": "noop"},
		CreatedAt:  time.Now(),
		Coordinate: TemporalCoordinate{LogicalTick: 0, WallClock: time.Now(), Epoch: "Phase84"},
	}

	snapshot := IdentityCrystalSnapshot{
		CrystalID:  crystal.ID,
		Identity:   req.Identity,
		Coordinate: crystal.Coordinate,
		Phase:      crystal.Phase,
		Stability:  crystal.Stability,
		Metadata:   map[string]string{"snapshot": "noop"},
	}

	metric := CrystallizationMetric{
		QualityScore:     1.0,
		PreStability:     1.0,
		PostStability:    1.0,
		CompressionRatio: 1.0,
		Tags:             []string{"noop", "crystallized"},
	}

	return &CrystallizationResult{
		Identity:           req.Identity,
		Metric:             metric,
		Crystal:            crystal,
		Snapshot:           snapshot,
		RecommendedActions: nil,
	}, nil
}

type NoopCrystallizationPlanner struct{}

func NewNoopCrystallizationPlanner() *NoopCrystallizationPlanner {
	return &NoopCrystallizationPlanner{}
}

func (p *NoopCrystallizationPlanner) BuildPlan(ctx context.Context, result *CrystallizationResult) (*CrystalPlan, error) {
	plan := &CrystalPlan{
		Identity:   result.Identity,
		Actions:    nil,
		Metric:     result.Metric,
		Metadata:   map[string]string{"planner": "noop"},
		CreatedAt:  time.Now(),
		Coordinate: result.Snapshot.Coordinate,
		Version:    "0.0.1-noop",
	}
	return plan, nil
}

type NoopCrystallizationExecutor struct{}

func NewNoopCrystallizationExecutor() *NoopCrystallizationExecutor {
	return &NoopCrystallizationExecutor{}
}

func (e *NoopCrystallizationExecutor) ExecutePlan(ctx context.Context, plan *CrystalPlan) error {
	return nil
}
