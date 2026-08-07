package main

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// Core identifiers & enums
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshIdentityPersonaCharterID uniquely identifies an identity charter.
type HarmonicSovereignMeshIdentityPersonaCharterID string

// HarmonicSovereignMeshIdentityPersonaTraitID uniquely identifies a persona trait.
type HarmonicSovereignMeshIdentityPersonaTraitID string

// HarmonicSovereignMeshIdentityPersonaDriftType describes identity drift categories.
type HarmonicSovereignMeshIdentityPersonaDriftType string

const (
	HarmonicSovereignMeshIdentityPersonaDriftTypeGovernance   HarmonicSovereignMeshIdentityPersonaDriftType = "governance_drift"
	HarmonicSovereignMeshIdentityPersonaDriftTypeTeleology    HarmonicSovereignMeshIdentityPersonaDriftType = "teleology_drift"
	HarmonicSovereignMeshIdentityPersonaDriftTypeEvolution    HarmonicSovereignMeshIdentityPersonaDriftType = "evolution_drift"
	HarmonicSovereignMeshIdentityPersonaDriftTypeConstitution HarmonicSovereignMeshIdentityPersonaDriftType = "constitution_drift"
)

// -----------------------------------------------------------------------------
// Config & dependencies
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshIdentityPersonaConfig configures THMS‑IP behavior.
type HarmonicSovereignMeshIdentityPersonaConfig struct {
	LoopTickInterval       time.Duration
	MaxAllowedDrift        float64
	ReinforcementIntensity float64
}

// HarmonicSovereignMeshIdentityPersonaDependencies wires THMS‑IP into Phases 111–116.
type HarmonicSovereignMeshIdentityPersonaDependencies struct {
	GovernanceMetricsProvider     HarmonicSovereignMeshIdentityPersonaGovernanceMetricsProvider
	TeleologicalMetricsProvider   HarmonicSovereignMeshIdentityPersonaTeleologicalMetricsProvider
	ConstitutionalMetricsProvider HarmonicSovereignMeshIdentityPersonaConstitutionalMetricsProvider
	IdentityCharterStore          HarmonicSovereignMeshIdentityPersonaCharterStore
	MetricsSink                   HarmonicSovereignMeshIdentityPersonaMetricsSink
}

// -----------------------------------------------------------------------------
// External metric providers
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshIdentityPersonaGovernanceMetricsProvider interface {
	GetGovernanceMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshIdentityPersonaTeleologicalMetricsProvider interface {
	GetTeleologicalMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshIdentityPersonaConstitutionalMetricsProvider interface {
	GetConstitutionalMetrics(ctx context.Context) (map[string]any, error)
}

// -----------------------------------------------------------------------------
// Identity & persona model
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshIdentityPersonaTrait struct {
	ID        HarmonicSovereignMeshIdentityPersonaTraitID
	Name      string
	Value     float64
	Locked    bool
	UpdatedAt time.Time
}

type HarmonicSovereignMeshIdentityPersonaCharter struct {
	ID        HarmonicSovereignMeshIdentityPersonaCharterID
	Traits    []*HarmonicSovereignMeshIdentityPersonaTrait
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool
}

type HarmonicSovereignMeshIdentityPersonaCharterStore interface {
	SaveCharter(ctx context.Context, charter *HarmonicSovereignMeshIdentityPersonaCharter) error
	LoadCharter(ctx context.Context, id HarmonicSovereignMeshIdentityPersonaCharterID) (*HarmonicSovereignMeshIdentityPersonaCharter, error)
	ListCharters(ctx context.Context) ([]*HarmonicSovereignMeshIdentityPersonaCharter, error)
}

// -----------------------------------------------------------------------------
// Persona kernel & decisions
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshIdentityPersonaKernel interface {
	EvaluateIdentity(
		ctx context.Context,
		charter *HarmonicSovereignMeshIdentityPersonaCharter,
		govMetrics map[string]any,
		telMetrics map[string]any,
		conMetrics map[string]any,
		cfg HarmonicSovereignMeshIdentityPersonaConfig,
	) (*HarmonicSovereignMeshIdentityPersonaDecision, error)
}

type HarmonicSovereignMeshIdentityPersonaDecision struct {
	CharterID        HarmonicSovereignMeshIdentityPersonaCharterID
	DriftType        HarmonicSovereignMeshIdentityPersonaDriftType
	ReinforcedTraits []string
	UpdatedCharter   *HarmonicSovereignMeshIdentityPersonaCharter
	DriftMagnitude   float64
	PersonaID        string
	IdentityVector   []float64
	ConfidenceScore  float64
	RiskScore        float64
	ExpectedValue    float64
	DecisionTag      string
	TimestampUnix    int64
}

// -----------------------------------------------------------------------------
// Metrics
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshIdentityPersonaMetrics struct {
	Timestamp        time.Time
	ActiveCharters   int
	ReinforcedTraits int
	DriftDetected    bool
	DriftMagnitude   float64
	Tags             map[string]string
}

type HarmonicSovereignMeshIdentityPersonaMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshIdentityPersonaMetrics) error
}

// -----------------------------------------------------------------------------
// Identity & Persona decision engine definitions
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshIdentityPersonaInput struct {
	Charter *HarmonicSovereignMeshIdentityPersonaCharter
	Metrics map[string]any
}

type HarmonicSovereignMeshIdentityPersonaIdentity struct {
	ID        string
	Signature string
}

type HarmonicSovereignMeshIdentityPersonaDecisionEngine interface {
	ComputeDecision(input *HarmonicSovereignMeshIdentityPersonaInput) (*HarmonicSovereignMeshIdentityPersonaDecision, error)
	ScoreIdentity(identity *HarmonicSovereignMeshIdentityPersonaIdentity) (float64, error)
	ScorePersona(persona *HarmonicSovereignMeshIdentityPersonaKernel) (float64, error)
	EvaluateRisk(input *HarmonicSovereignMeshIdentityPersonaInput) (float64, error)
	EvaluateExpectedValue(input *HarmonicSovereignMeshIdentityPersonaInput) (float64, error)
}

type HarmonicSovereignMeshIdentityPersonaIdentityEngine interface {
	EvaluateIdentityContinuity(ctx context.Context, charter *HarmonicSovereignMeshIdentityPersonaCharter) error
}

type HarmonicSovereignMeshIdentityPersonaKernelEngine interface {
	ProcessKernel(ctx context.Context, k HarmonicSovereignMeshIdentityPersonaKernel) error
}

type HarmonicSovereignMeshIdentityPersonaMetricsEngine interface {
	ProcessMetrics(ctx context.Context, m HarmonicSovereignMeshIdentityPersonaMetrics) error
}

type HarmonicSovereignMeshIdentityPersonaLoopEngine interface {
	RunLoop(ctx context.Context) error
}

type HarmonicSovereignMeshIdentityPersonaDecisionDependencies struct {
	IdentityEngine HarmonicSovereignMeshIdentityPersonaIdentityEngine
	PersonaEngine  HarmonicSovereignMeshIdentityPersonaKernelEngine
	MetricsEngine  HarmonicSovereignMeshIdentityPersonaMetricsEngine
	LoopEngine     HarmonicSovereignMeshIdentityPersonaLoopEngine
}

type HarmonicSovereignMeshIdentityPersonaDefaultDecisionEngine struct {
	Dependencies *HarmonicSovereignMeshIdentityPersonaDecisionDependencies
}

func NewHarmonicSovereignMeshIdentityPersonaDefaultDecisionEngine(
	deps *HarmonicSovereignMeshIdentityPersonaDecisionDependencies,
) *HarmonicSovereignMeshIdentityPersonaDefaultDecisionEngine {
	return &HarmonicSovereignMeshIdentityPersonaDefaultDecisionEngine{
		Dependencies: deps,
	}
}

// -----------------------------------------------------------------------------
// Default execution engine (THMS-IP)
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshIdentityPersonaEngine interface {
	Run(ctx context.Context) error
}

type HarmonicSovereignMeshIdentityPersonaDefaultEngine struct {
	cfg    HarmonicSovereignMeshIdentityPersonaConfig
	deps   HarmonicSovereignMeshIdentityPersonaDependencies
	kernel HarmonicSovereignMeshIdentityPersonaKernel
}

func NewHarmonicSovereignMeshIdentityPersonaDefaultEngine(
	cfg HarmonicSovereignMeshIdentityPersonaConfig,
	deps HarmonicSovereignMeshIdentityPersonaDependencies,
	kernel HarmonicSovereignMeshIdentityPersonaKernel,
) *HarmonicSovereignMeshIdentityPersonaDefaultEngine {
	return &HarmonicSovereignMeshIdentityPersonaDefaultEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
	}
}

func (e *HarmonicSovereignMeshIdentityPersonaDefaultEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			charters, err := e.deps.IdentityCharterStore.ListCharters(ctx)
			if err != nil {
				continue
			}

			govMetrics, _ := e.deps.GovernanceMetricsProvider.GetGovernanceMetrics(ctx)
			telMetrics, _ := e.deps.TeleologicalMetricsProvider.GetTeleologicalMetrics(ctx)
			conMetrics, _ := e.deps.ConstitutionalMetricsProvider.GetConstitutionalMetrics(ctx)

			metrics := HarmonicSovereignMeshIdentityPersonaMetrics{
				Timestamp:      time.Now(),
				Tags:           map[string]string{"engine": "HarmonicSovereignMeshIdentityPersonaDefaultEngine"},
				ActiveCharters: len(charters),
			}

			for _, charter := range charters {
				if !charter.Active {
					continue
				}

				decision, err := e.kernel.EvaluateIdentity(
					ctx, charter, govMetrics, telMetrics, conMetrics, e.cfg,
				)
				if err != nil || decision == nil {
					continue
				}

				if decision.UpdatedCharter != nil {
					_ = e.deps.IdentityCharterStore.SaveCharter(ctx, decision.UpdatedCharter)
				}

				metrics.ReinforcedTraits += len(decision.ReinforcedTraits)
				metrics.DriftDetected = true
				metrics.DriftMagnitude = decision.DriftMagnitude
			}

			_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
		}
	}
}
