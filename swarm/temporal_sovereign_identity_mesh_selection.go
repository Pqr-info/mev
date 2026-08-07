package main

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// Core identifiers & enums
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSelectionLineageFitness represents the computed fitness score.
type HarmonicSovereignMeshSelectionLineageFitness float64

// HarmonicSovereignMeshSelectionAction describes what the selection kernel decides.
type HarmonicSovereignMeshSelectionAction string

const (
	HarmonicSovereignMeshSelectionActionAmplify    HarmonicSovereignMeshSelectionAction = "amplify"
	HarmonicSovereignMeshSelectionActionSustain    HarmonicSovereignMeshSelectionAction = "sustain"
	HarmonicSovereignMeshSelectionActionMerge      HarmonicSovereignMeshSelectionAction = "merge"
	HarmonicSovereignMeshSelectionActionExtinguish HarmonicSovereignMeshSelectionAction = "extinguish"
)

// -----------------------------------------------------------------------------
// Config & dependencies
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSelectionConfig configures THMS‑SEL behavior.
type HarmonicSovereignMeshSelectionConfig struct {
	LoopTickInterval    time.Duration
	AmplifyThreshold    float64
	SustainThreshold    float64
	MergeThreshold      float64
	ExtinguishThreshold float64
}

// HarmonicSovereignMeshSelectionDependencies wires THMS‑SEL into Phase 111.
type HarmonicSovereignMeshSelectionDependencies struct {
	SpeciationLattice HarmonicSovereignMeshSelectionSpeciationLatticeProvider
	MetricsSink       HarmonicSovereignMeshSelectionMetricsSink
}

// -----------------------------------------------------------------------------
// External interfaces to Phase 111
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshSelectionLineageState struct {
	ID      string
	Fitness float64
}

// HarmonicSovereignMeshSelectionLatticeSnapshot is a read‑only view of the speciation lattice.
type HarmonicSovereignMeshSelectionLatticeSnapshot struct {
	Lineages map[string]*HarmonicSovereignMeshSelectionLineageState
}

// HarmonicSovereignMeshSelectionSpeciationLatticeProvider exposes the Phase 111 lattice.
type HarmonicSovereignMeshSelectionSpeciationLatticeProvider interface {
	Lattice(ctx context.Context) (*HarmonicSovereignMeshSelectionLatticeSnapshot, error)
	ApplySelection(ctx context.Context, decision *HarmonicSovereignMeshSelectionDecision) error
}

// HarmonicSovereignMeshSelectionKernel evaluates lineage fitness and decides actions.
type HarmonicSovereignMeshSelectionKernel interface {
	EvaluateFitness(
		ctx context.Context,
		lineage *HarmonicSovereignMeshSelectionLineageState,
	) (HarmonicSovereignMeshSelectionLineageFitness, error)

	Decide(
		ctx context.Context,
		lineage *HarmonicSovereignMeshSelectionLineageState,
		fitness HarmonicSovereignMeshSelectionLineageFitness,
	) (HarmonicSovereignMeshSelectionAction, error)
}

// HarmonicSovereignMeshSelectionDecision describes the actions to apply to the lattice.
type HarmonicSovereignMeshSelectionDecision struct {
	AmplifyLineages    []string
	SustainLineages    []string
	MergeLineages      []string
	ExtinguishLineages []string
}

// -----------------------------------------------------------------------------
// Selection loop
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSelectionLoop drives continuous selection.
type HarmonicSovereignMeshSelectionLoop interface {
	Start(ctx context.Context) error
}

// -----------------------------------------------------------------------------
// Metrics
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSelectionMetrics captures observability for THMS‑SEL.
type HarmonicSovereignMeshSelectionMetrics struct {
	Timestamp         time.Time
	AmplifiedCount    int
	SustainedCount    int
	MergedCount       int
	ExtinguishedCount int
	AverageFitness    float64
	Tags              map[string]string
}

// HarmonicSovereignMeshSelectionMetricsSink receives metrics.
type HarmonicSovereignMeshSelectionMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshSelectionMetrics) error
}

// -----------------------------------------------------------------------------
// Selection engine (default implementation)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshSelectionEngine is the main interface for Phase 112.
type HarmonicSovereignMeshSelectionEngine interface {
	Run(ctx context.Context) error
}

// HarmonicSovereignMeshSelectionDefaultEngine is the default THMS‑SEL implementation.
type HarmonicSovereignMeshSelectionDefaultEngine struct {
	cfg    HarmonicSovereignMeshSelectionConfig
	deps   HarmonicSovereignMeshSelectionDependencies
	kernel HarmonicSovereignMeshSelectionKernel
}

// NewHarmonicSovereignMeshSelectionDefaultEngine constructs the default engine.
func NewHarmonicSovereignMeshSelectionDefaultEngine(
	cfg HarmonicSovereignMeshSelectionConfig,
	deps HarmonicSovereignMeshSelectionDependencies,
	kernel HarmonicSovereignMeshSelectionKernel,
) *HarmonicSovereignMeshSelectionDefaultEngine {
	return &HarmonicSovereignMeshSelectionDefaultEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
	}
}

// Run executes the selection loop.
func (e *HarmonicSovereignMeshSelectionDefaultEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			snapshot, err := e.deps.SpeciationLattice.Lattice(ctx)
			if err != nil || snapshot == nil {
				continue
			}

			decision := &HarmonicSovereignMeshSelectionDecision{}
			var fitnessSum float64
			var fitnessCount int

			for _, lineage := range snapshot.Lineages {
				fitness, err := e.kernel.EvaluateFitness(ctx, lineage)
				if err != nil {
					continue
				}

				action, err := e.kernel.Decide(ctx, lineage, fitness)
				if err != nil {
					continue
				}

				switch action {
				case HarmonicSovereignMeshSelectionActionAmplify:
					decision.AmplifyLineages = append(decision.AmplifyLineages, lineage.ID)
				case HarmonicSovereignMeshSelectionActionSustain:
					decision.SustainLineages = append(decision.SustainLineages, lineage.ID)
				case HarmonicSovereignMeshSelectionActionMerge:
					decision.MergeLineages = append(decision.MergeLineages, lineage.ID)
				case HarmonicSovereignMeshSelectionActionExtinguish:
					decision.ExtinguishLineages = append(decision.ExtinguishLineages, lineage.ID)
				}

				fitnessSum += float64(fitness)
				fitnessCount++
			}

			_ = e.deps.SpeciationLattice.ApplySelection(ctx, decision)

			var avgFitness float64
			if fitnessCount > 0 {
				avgFitness = fitnessSum / float64(fitnessCount)
			}

			metrics := HarmonicSovereignMeshSelectionMetrics{
				Timestamp:         time.Now(),
				AmplifiedCount:    len(decision.AmplifyLineages),
				SustainedCount:    len(decision.SustainLineages),
				MergedCount:       len(decision.MergeLineages),
				ExtinguishedCount: len(decision.ExtinguishLineages),
				AverageFitness:    avgFitness,
				Tags: map[string]string{
					"engine": "HarmonicSovereignMeshSelectionDefaultEngine",
				},
			}

			_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
		}
	}
}
