package main

import (
	"context"
	"time"
)

type HarmonicSovereignMeshSpeciationLineageID string
type HarmonicSovereignMeshSpeciationForkReason string

// HarmonicSovereignMeshSpeciationConfig configures the speciation engine.
type HarmonicSovereignMeshSpeciationConfig struct {
	MaxConcurrentLineages int
	LoopTickInterval      time.Duration
}

// HarmonicSovereignMeshSpeciationDependencies wires THMS-E into the broader organism.
type HarmonicSovereignMeshSpeciationDependencies struct {
	// EvolutionStream provides the Phase 110 Temporal Harmonic Mesh Evolution Engine output.
	EvolutionStream HarmonicSovereignMeshSpeciationEvolutionStream

	// MetricsSink receives speciation metrics for observability.
	MetricsSink HarmonicSovereignMeshSpeciationMetricsSink

	// LineageStore persists lineage state and history.
	LineageStore HarmonicSovereignMeshSpeciationLineageStore
}

// -----------------------------------------------------------------------------
// External interfaces (to Phase 110 and observability)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSpeciationEvolutionSnapshot represents a single temporal harmonic snapshot
// emitted by the Phase 110 Evolution Engine.
type HarmonicSovereignMeshSpeciationEvolutionSnapshot struct {
	Timestamp time.Time

	// HarmonicState is an opaque representation of the mesh harmonic field.
	HarmonicState any

	// Optional metadata for routing/speciation decisions.
	Metadata map[string]any
}

// HarmonicSovereignMeshSpeciationEvolutionStream is the interface to consume evolution snapshots.
type HarmonicSovereignMeshSpeciationEvolutionStream interface {
	// NextSnapshot blocks or polls for the next evolution snapshot.
	NextSnapshot(ctx context.Context) (*HarmonicSovereignMeshSpeciationEvolutionSnapshot, error)
}

// HarmonicSovereignMeshSpeciationMetricsSink receives metrics emitted by the speciation engine.
type HarmonicSovereignMeshSpeciationMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshSpeciationMetrics) error
}

// HarmonicSovereignMeshSpeciationLineageStore persists and retrieves lineage state.
type HarmonicSovereignMeshSpeciationLineageStore interface {
	SaveLineage(ctx context.Context, lineage *HarmonicSovereignMeshSpeciationLineage) error
	LoadLineage(ctx context.Context, id HarmonicSovereignMeshSpeciationLineageID) (*HarmonicSovereignMeshSpeciationLineage, error)
	ListLineages(ctx context.Context) ([]*HarmonicSovereignMeshSpeciationLineage, error)
}

// -----------------------------------------------------------------------------
// Speciation lattice & kernel
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSpeciationLattice represents the active ecosystem of lineages.
type HarmonicSovereignMeshSpeciationLattice struct {
	Lineages map[HarmonicSovereignMeshSpeciationLineageID]*HarmonicSovereignMeshSpeciationLineage
}

// HarmonicSovereignMeshSpeciationLineage represents a single temporal harmonic species/branch.
type HarmonicSovereignMeshSpeciationLineage struct {
	ID HarmonicSovereignMeshSpeciationLineageID

	// Current harmonic state associated with this lineage.
	CurrentState any

	// Historical divergence metrics and ancestry.
	ParentID      *HarmonicSovereignMeshSpeciationLineageID
	ForkReason    HarmonicSovereignMeshSpeciationForkReason
	CreatedAt     time.Time
	LastUpdatedAt time.Time

	// Arbitrary lineage-local metadata.
	Metadata map[string]any
}

// HarmonicSovereignMeshSpeciationKernel encapsulates the rules for speciation, merging, and pruning.
type HarmonicSovereignMeshSpeciationKernel interface {
	// EvaluateSnapshot decides how a new evolution snapshot affects the lattice.
	EvaluateSnapshot(
		ctx context.Context,
		lattice *HarmonicSovereignMeshSpeciationLattice,
		snapshot *HarmonicSovereignMeshSpeciationEvolutionSnapshot,
	) (*HarmonicSovereignMeshSpeciationKernelDecision, error)
}

// HarmonicSovereignMeshSpeciationKernelDecision describes the actions to apply to the lattice.
type HarmonicSovereignMeshSpeciationKernelDecision struct {
	// NewLineages to be created as a result of speciation.
	NewLineages []*HarmonicSovereignMeshSpeciationLineage

	// UpdatedLineages whose state should be mutated.
	UpdatedLineages []*HarmonicSovereignMeshSpeciationLineage

	// LineagesToPrune that should be removed from the lattice.
	LineagesToPrune []HarmonicSovereignMeshSpeciationLineageID
}

// -----------------------------------------------------------------------------
// Speciation loop & metrics
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSpeciationLoop drives the continuous speciation process.
type HarmonicSovereignMeshSpeciationLoop interface {
	// Start begins the speciation loop and blocks until context cancellation or fatal error.
	Start(ctx context.Context) error
}

// HarmonicSovereignMeshSpeciationMetrics captures high-level observability for THMS‑E.
type HarmonicSovereignMeshSpeciationMetrics struct {
	Timestamp          time.Time
	ActiveLineages     int
	CreatedLineages    int
	PrunedLineages     int
	AverageDivergence  float64
	AverageConvergence float64
	Tags               map[string]string
}

// -----------------------------------------------------------------------------
// Speciation engine (default implementation)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSpeciationEngine is the main interface for Phase 111.
type HarmonicSovereignMeshSpeciationEngine interface {
	Run(ctx context.Context) error
	Lattice() *HarmonicSovereignMeshSpeciationLattice
}

// HarmonicSovereignMeshSpeciationDefaultEngine is the default THMS‑E implementation.
type HarmonicSovereignMeshSpeciationDefaultEngine struct {
	cfg     HarmonicSovereignMeshSpeciationConfig
	deps    HarmonicSovereignMeshSpeciationDependencies
	lattice *HarmonicSovereignMeshSpeciationLattice
	kernel  HarmonicSovereignMeshSpeciationKernel
}

// NewHarmonicSovereignMeshSpeciationDefaultEngine constructs the default engine.
func NewHarmonicSovereignMeshSpeciationDefaultEngine(
	cfg HarmonicSovereignMeshSpeciationConfig,
	deps HarmonicSovereignMeshSpeciationDependencies,
	kernel HarmonicSovereignMeshSpeciationKernel,
) *HarmonicSovereignMeshSpeciationDefaultEngine {
	return &HarmonicSovereignMeshSpeciationDefaultEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
		lattice: &HarmonicSovereignMeshSpeciationLattice{
			Lineages: make(map[HarmonicSovereignMeshSpeciationLineageID]*HarmonicSovereignMeshSpeciationLineage),
		},
	}
}

// Run executes the speciation loop.
func (e *HarmonicSovereignMeshSpeciationDefaultEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			snapshot, err := e.deps.EvolutionStream.NextSnapshot(ctx)
			if err != nil || snapshot == nil {
				continue
			}

			decision, err := e.kernel.EvaluateSnapshot(ctx, e.lattice, snapshot)
			if err != nil {
				continue
			}

			e.applyKernelDecision(ctx, decision)
		}
	}
}

// Lattice returns the current lattice.
func (e *HarmonicSovereignMeshSpeciationDefaultEngine) Lattice() *HarmonicSovereignMeshSpeciationLattice {
	return e.lattice
}

// applyKernelDecision mutates the lattice and emits metrics.
func (e *HarmonicSovereignMeshSpeciationDefaultEngine) applyKernelDecision(
	ctx context.Context,
	decision *HarmonicSovereignMeshSpeciationKernelDecision,
) {
	for _, lineage := range decision.NewLineages {
		e.lattice.Lineages[lineage.ID] = lineage
	}
	for _, lineage := range decision.UpdatedLineages {
		e.lattice.Lineages[lineage.ID] = lineage
	}
	for _, id := range decision.LineagesToPrune {
		delete(e.lattice.Lineages, id)
	}

	metrics := HarmonicSovereignMeshSpeciationMetrics{
		Timestamp:       time.Now(),
		ActiveLineages:  len(e.lattice.Lineages),
		CreatedLineages: len(decision.NewLineages),
		PrunedLineages:  len(decision.LineagesToPrune),
		Tags: map[string]string{
			"engine": "HarmonicSovereignMeshSpeciationDefaultEngine",
		},
	}

	_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
}
