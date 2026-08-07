package main

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// Core identifiers & enums
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalAttractorID uniquely identifies a teleological attractor.
type HarmonicSovereignMeshTeleologicalAttractorID string

// HarmonicSovereignMeshTeleologicalTrajectoryID uniquely identifies a temporal trajectory.
type HarmonicSovereignMeshTeleologicalTrajectoryID string

// HarmonicSovereignMeshTeleologicalDriveAction describes the drive kernel's decision.
type HarmonicSovereignMeshTeleologicalDriveAction string

const (
	HarmonicSovereignMeshTeleologicalDriveActionPursue   HarmonicSovereignMeshTeleologicalDriveAction = "pursue"
	HarmonicSovereignMeshTeleologicalDriveActionAbandon  HarmonicSovereignMeshTeleologicalDriveAction = "abandon"
	HarmonicSovereignMeshTeleologicalDriveActionSwitch   HarmonicSovereignMeshTeleologicalDriveAction = "switch"
	HarmonicSovereignMeshTeleologicalDriveActionGenerate HarmonicSovereignMeshTeleologicalDriveAction = "generate"
)

// -----------------------------------------------------------------------------
// Config & dependencies
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalConfig configures THMS‑TD behavior.
type HarmonicSovereignMeshTeleologicalConfig struct {
	LoopTickInterval     time.Duration
	AttractorSensitivity float64
	TrajectoryHorizon    int
	MaxActiveAttractors  int
}

// HarmonicSovereignMeshTeleologicalDependencies wires THMS‑TD into Phases 110–114.
type HarmonicSovereignMeshTeleologicalDependencies struct {
	GovernanceMetricsProvider HarmonicSovereignMeshTeleologicalGovernanceMetricsProvider
	SelectionMetricsProvider  HarmonicSovereignMeshTeleologicalSelectionMetricsProvider
	SpeciationMetricsProvider HarmonicSovereignMeshTeleologicalSpeciationMetricsProvider
	AttractorStore            HarmonicSovereignMeshTeleologicalAttractorStore
	TrajectoryStore           HarmonicSovereignMeshTeleologicalTrajectoryStore
	MetricsSink               HarmonicSovereignMeshTeleologicalMetricsSink
}

// -----------------------------------------------------------------------------
// External metric providers (Phases 111–114)
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshTeleologicalGovernanceMetricsProvider interface {
	GetGovernanceMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshTeleologicalSelectionMetricsProvider interface {
	GetSelectionMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshTeleologicalSpeciationMetricsProvider interface {
	GetSpeciationMetrics(ctx context.Context) (map[string]any, error)
}

// -----------------------------------------------------------------------------
// Teleological attractor model
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshTeleologicalAttractor struct {
	ID        HarmonicSovereignMeshTeleologicalAttractorID
	Signature map[string]any
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool
}

// HarmonicSovereignMeshTeleologicalAttractorStore persists attractors.
type HarmonicSovereignMeshTeleologicalAttractorStore interface {
	SaveAttractor(
		ctx context.Context,
		attractor *HarmonicSovereignMeshTeleologicalAttractor,
	) error

	LoadAttractor(
		ctx context.Context,
		id HarmonicSovereignMeshTeleologicalAttractorID,
	) (*HarmonicSovereignMeshTeleologicalAttractor, error)

	ListAttractors(
		ctx context.Context,
	) ([]*HarmonicSovereignMeshTeleologicalAttractor, error)
}

// -----------------------------------------------------------------------------
// Teleological trajectory model
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalTrajectory represents a temporal path toward an attractor.
type HarmonicSovereignMeshTeleologicalTrajectory struct {
	ID        HarmonicSovereignMeshTeleologicalTrajectoryID
	Attractor HarmonicSovereignMeshTeleologicalAttractorID
	Steps     []map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	Completed bool
}

// HarmonicSovereignMeshTeleologicalTrajectoryStore persists trajectories.
type HarmonicSovereignMeshTeleologicalTrajectoryStore interface {
	SaveTrajectory(
		ctx context.Context,
		traj *HarmonicSovereignMeshTeleologicalTrajectory,
	) error

	LoadTrajectory(
		ctx context.Context,
		id HarmonicSovereignMeshTeleologicalTrajectoryID,
	) (*HarmonicSovereignMeshTeleologicalTrajectory, error)

	ListTrajectories(
		ctx context.Context,
	) ([]*HarmonicSovereignMeshTeleologicalTrajectory, error)
}

// -----------------------------------------------------------------------------
// Teleological drive kernel
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalDriveKernel decides attractor pursuit and trajectory generation.
type HarmonicSovereignMeshTeleologicalDriveKernel interface {
	EvaluateDrive(
		ctx context.Context,
		attractors []*HarmonicSovereignMeshTeleologicalAttractor,
		govMetrics map[string]any,
		selMetrics map[string]any,
		specMetrics map[string]any,
		cfg HarmonicSovereignMeshTeleologicalConfig,
	) (*HarmonicSovereignMeshTeleologicalDriveDecision, error)
}

// HarmonicSovereignMeshTeleologicalDriveDecision describes teleological actions.
type HarmonicSovereignMeshTeleologicalDriveDecision struct {
	Action          HarmonicSovereignMeshTeleologicalDriveAction
	TargetAttractor HarmonicSovereignMeshTeleologicalAttractorID
	NewAttractor    *HarmonicSovereignMeshTeleologicalAttractor
	NewTrajectory   *HarmonicSovereignMeshTeleologicalTrajectory
}

// -----------------------------------------------------------------------------
// Teleological loop
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalLoop drives continuous teleological evaluation.
type HarmonicSovereignMeshTeleologicalLoop interface {
	Start(ctx context.Context) error
}

// -----------------------------------------------------------------------------
// Metrics
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalMetrics captures observability for THMS‑TD.
type HarmonicSovereignMeshTeleologicalMetrics struct {
	Timestamp             time.Time
	ActiveAttractors      int
	GeneratedAttractors   int
	CompletedTrajectories int
	DriveAction           string
	Tags                  map[string]string
}

// HarmonicSovereignMeshTeleologicalMetricsSink receives teleological metrics.
type HarmonicSovereignMeshTeleologicalMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshTeleologicalMetrics) error
}

// -----------------------------------------------------------------------------
// Teleological drive engine (default implementation)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshTeleologicalDriveEngine is the main interface for Phase 115.
type HarmonicSovereignMeshTeleologicalDriveEngine interface {
	Run(ctx context.Context) error
}

// HarmonicSovereignMeshTeleologicalDefaultDriveEngine is the default THMS‑TD implementation.
type HarmonicSovereignMeshTeleologicalDefaultDriveEngine struct {
	cfg    HarmonicSovereignMeshTeleologicalConfig
	deps   HarmonicSovereignMeshTeleologicalDependencies
	kernel HarmonicSovereignMeshTeleologicalDriveKernel
}

// NewHarmonicSovereignMeshTeleologicalDefaultDriveEngine constructs the default engine.
func NewHarmonicSovereignMeshTeleologicalDefaultDriveEngine(
	cfg HarmonicSovereignMeshTeleologicalConfig,
	deps HarmonicSovereignMeshTeleologicalDependencies,
	kernel HarmonicSovereignMeshTeleologicalDriveKernel,
) *HarmonicSovereignMeshTeleologicalDefaultDriveEngine {
	return &HarmonicSovereignMeshTeleologicalDefaultDriveEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
	}
}

// Run executes the teleological drive loop.
func (e *HarmonicSovereignMeshTeleologicalDefaultDriveEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			attractors, err := e.deps.AttractorStore.ListAttractors(ctx)
			if err != nil {
				continue
			}

			govMetrics, _ := e.deps.GovernanceMetricsProvider.GetGovernanceMetrics(ctx)
			selMetrics, _ := e.deps.SelectionMetricsProvider.GetSelectionMetrics(ctx)
			specMetrics, _ := e.deps.SpeciationMetricsProvider.GetSpeciationMetrics(ctx)

			decision, err := e.kernel.EvaluateDrive(
				ctx, attractors, govMetrics, selMetrics, specMetrics, e.cfg,
			)
			if err != nil || decision == nil {
				continue
			}

			metrics := HarmonicSovereignMeshTeleologicalMetrics{
				Timestamp:   time.Now(),
				Tags:        map[string]string{"engine": "HarmonicSovereignMeshTeleologicalDefaultDriveEngine"},
				DriveAction: string(decision.Action),
			}

			switch decision.Action {
			case HarmonicSovereignMeshTeleologicalDriveActionGenerate:
				if decision.NewAttractor != nil {
					_ = e.deps.AttractorStore.SaveAttractor(ctx, decision.NewAttractor)
					metrics.GeneratedAttractors++
				}

			case HarmonicSovereignMeshTeleologicalDriveActionPursue,
				HarmonicSovereignMeshTeleologicalDriveActionSwitch:
				if decision.NewTrajectory != nil {
					_ = e.deps.TrajectoryStore.SaveTrajectory(ctx, decision.NewTrajectory)
				}

			case HarmonicSovereignMeshTeleologicalDriveActionAbandon:
				// No persistence required.
			}

			_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
		}
	}
}
