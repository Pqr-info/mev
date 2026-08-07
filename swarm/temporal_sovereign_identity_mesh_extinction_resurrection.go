package main

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// Core identifiers & enums
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionLineageID identifies a lineage to extinguish or resurrect.
type HarmonicSovereignMeshExtinctionResurrectionLineageID string

// HarmonicSovereignMeshExtinctionResurrectionAction describes extinction/resurrection decisions.
type HarmonicSovereignMeshExtinctionResurrectionAction string

const (
	HarmonicSovereignMeshExtinctionResurrectionActionExtinguish HarmonicSovereignMeshExtinctionResurrectionAction = "extinguish"
	HarmonicSovereignMeshExtinctionResurrectionActionResurrect  HarmonicSovereignMeshExtinctionResurrectionAction = "resurrect"
)

// -----------------------------------------------------------------------------
// Config & dependencies
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionConfig configures THMS‑ER behavior.
type HarmonicSovereignMeshExtinctionResurrectionConfig struct {
	LoopTickInterval             time.Duration
	ExtinctionFitnessThreshold   float64
	ResurrectionFitnessThreshold float64
	MaxResurrectionCount         int
}

// HarmonicSovereignMeshExtinctionResurrectionDependencies wires THMS‑ER into Phase 111 & 112.
type HarmonicSovereignMeshExtinctionResurrectionDependencies struct {
	SelectionLattice HarmonicSovereignMeshExtinctionResurrectionSelectionLatticeProvider
	ArchiveStore     HarmonicSovereignMeshExtinctionResurrectionArchiveStore
	MetricsSink      HarmonicSovereignMeshExtinctionResurrectionMetricsSink
}

// -----------------------------------------------------------------------------
// External interfaces to Phase 111 & 112
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionSelectionLatticeProvider exposes the current lattice.
type HarmonicSovereignMeshExtinctionResurrectionSelectionLatticeProvider interface {
	Lattice(ctx context.Context) (*HarmonicSovereignMeshExtinctionResurrectionLatticeSnapshot, error)
	ApplyExtinctionResurrection(ctx context.Context, decision *HarmonicSovereignMeshExtinctionResurrectionDecision) error
}

// HarmonicSovereignMeshExtinctionResurrectionLatticeSnapshot is a read‑only view of the lineage ecosystem.
type HarmonicSovereignMeshExtinctionResurrectionLatticeSnapshot struct {
	Lineages map[HarmonicSovereignMeshExtinctionResurrectionLineageID]*HarmonicSovereignMeshExtinctionResurrectionLineageState
}

// HarmonicSovereignMeshExtinctionResurrectionLineageState represents lineage state relevant to ER decisions.
type HarmonicSovereignMeshExtinctionResurrectionLineageState struct {
	ID        HarmonicSovereignMeshExtinctionResurrectionLineageID
	State     any
	Fitness   float64
	Metadata  map[string]any
	CreatedAt time.Time
}

// HarmonicSovereignMeshExtinctionResurrectionArchiveStore persists historical lineage states.
type HarmonicSovereignMeshExtinctionResurrectionArchiveStore interface {
	SaveToArchive(
		ctx context.Context,
		lineage *HarmonicSovereignMeshExtinctionResurrectionLineageState,
	) error

	LoadFromArchive(
		ctx context.Context,
		id HarmonicSovereignMeshExtinctionResurrectionLineageID,
	) (*HarmonicSovereignMeshExtinctionResurrectionLineageState, error)

	ListArchived(
		ctx context.Context,
	) ([]*HarmonicSovereignMeshExtinctionResurrectionLineageState, error)
}

// -----------------------------------------------------------------------------
// Extinction & Resurrection kernel
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionKernel decides extinction/resurrection actions.
type HarmonicSovereignMeshExtinctionResurrectionKernel interface {
	Evaluate(
		ctx context.Context,
		lineage *HarmonicSovereignMeshExtinctionResurrectionLineageState,
		cfg HarmonicSovereignMeshExtinctionResurrectionConfig,
	) (HarmonicSovereignMeshExtinctionResurrectionAction, error)
}

// HarmonicSovereignMeshExtinctionResurrectionDecision describes actions to apply to the lattice.
type HarmonicSovereignMeshExtinctionResurrectionDecision struct {
	LineagesToExtinguish []HarmonicSovereignMeshExtinctionResurrectionLineageID
	LineagesToResurrect  []HarmonicSovereignMeshExtinctionResurrectionLineageID
}

// -----------------------------------------------------------------------------
// Extinction & Resurrection loop
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionLoop drives continuous ER processing.
type HarmonicSovereignMeshExtinctionResurrectionLoop interface {
	Start(ctx context.Context) error
}

// -----------------------------------------------------------------------------
// Metrics
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionMetrics captures observability for THMS‑ER.
type HarmonicSovereignMeshExtinctionResurrectionMetrics struct {
	Timestamp         time.Time
	ExtinguishedCount int
	ResurrectedCount  int
	AverageFitness    float64
	Tags              map[string]string
}

// HarmonicSovereignMeshExtinctionResurrectionMetricsSink receives metrics.
type HarmonicSovereignMeshExtinctionResurrectionMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshExtinctionResurrectionMetrics) error
}

// -----------------------------------------------------------------------------
// Extinction & Resurrection engine (default implementation)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshExtinctionResurrectionEngine is the main interface for Phase 113.
type HarmonicSovereignMeshExtinctionResurrectionEngine interface {
	Run(ctx context.Context) error
}

type HarmonicSovereignMeshExtinctionResurrectionDefaultEngine struct {
	cfg    HarmonicSovereignMeshExtinctionResurrectionConfig
	deps   HarmonicSovereignMeshExtinctionResurrectionDependencies
	kernel HarmonicSovereignMeshExtinctionResurrectionKernel
}

func NewHarmonicSovereignMeshExtinctionResurrectionDefaultEngine(
	cfg HarmonicSovereignMeshExtinctionResurrectionConfig,
	deps HarmonicSovereignMeshExtinctionResurrectionDependencies,
	kernel HarmonicSovereignMeshExtinctionResurrectionKernel,
) *HarmonicSovereignMeshExtinctionResurrectionDefaultEngine {
	return &HarmonicSovereignMeshExtinctionResurrectionDefaultEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
	}
}

func (e *HarmonicSovereignMeshExtinctionResurrectionDefaultEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			snapshot, err := e.deps.SelectionLattice.Lattice(ctx)
			if err != nil || snapshot == nil {
				continue
			}

			decision := &HarmonicSovereignMeshExtinctionResurrectionDecision{}
			var fitnessSum float64
			var fitnessCount int

			for _, lineage := range snapshot.Lineages {
				action, err := e.kernel.Evaluate(ctx, lineage, e.cfg)
				if err != nil {
					continue
				}

				switch action {
				case HarmonicSovereignMeshExtinctionResurrectionActionExtinguish:
					decision.LineagesToExtinguish = append(decision.LineagesToExtinguish, lineage.ID)
					_ = e.deps.ArchiveStore.SaveToArchive(ctx, lineage)

				case HarmonicSovereignMeshExtinctionResurrectionActionResurrect:
					decision.LineagesToResurrect = append(decision.LineagesToResurrect, lineage.ID)
				}

				fitnessSum += lineage.Fitness
				fitnessCount++
			}

			_ = e.deps.SelectionLattice.ApplyExtinctionResurrection(ctx, decision)

			metrics := HarmonicSovereignMeshExtinctionResurrectionMetrics{
				Timestamp:         time.Now(),
				ExtinguishedCount: len(decision.LineagesToExtinguish),
				ResurrectedCount:  len(decision.LineagesToResurrect),
				AverageFitness: func() float64 {
					if fitnessCount == 0 {
						return 0
					}
					return fitnessSum / float64(fitnessCount)
				}(),
				Tags: map[string]string{"engine": "HarmonicSovereignMeshExtinctionResurrectionDefaultEngine"},
			}

			_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
		}
	}
}
