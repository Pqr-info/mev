package main

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// Core identifiers & enums
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalCharterID uniquely identifies a constitutional charter.
type HarmonicSovereignMeshConstitutionalCharterID string

// HarmonicSovereignMeshConstitutionalClauseID uniquely identifies a constitutional clause.
type HarmonicSovereignMeshConstitutionalClauseID string

// HarmonicSovereignMeshConstitutionalAmendmentType describes how a clause mutates.
type HarmonicSovereignMeshConstitutionalAmendmentType string

const (
	HarmonicSovereignMeshConstitutionalAmendmentTypeAddClause    HarmonicSovereignMeshConstitutionalAmendmentType = "add_clause"
	HarmonicSovereignMeshConstitutionalAmendmentTypeUpdateClause HarmonicSovereignMeshConstitutionalAmendmentType = "update_clause"
	HarmonicSovereignMeshConstitutionalAmendmentTypeRemoveClause HarmonicSovereignMeshConstitutionalAmendmentType = "remove_clause"
	HarmonicSovereignMeshConstitutionalAmendmentTypeLockClause   HarmonicSovereignMeshConstitutionalAmendmentType = "lock_clause"
)

// -----------------------------------------------------------------------------
// Config & dependencies
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalConfig configures THMS‑CE behavior.
type HarmonicSovereignMeshConstitutionalConfig struct {
	LoopTickInterval            time.Duration
	AmendmentConsensusThreshold float64
	MaxActiveCharters           int
}

// HarmonicSovereignMeshConstitutionalDependencies wires THMS‑CE into Phases 111–115.
type HarmonicSovereignMeshConstitutionalDependencies struct {
	GovernanceMetricsProvider   HarmonicSovereignMeshConstitutionalGovernanceMetricsProvider
	TeleologicalMetricsProvider HarmonicSovereignMeshConstitutionalTeleologicalMetricsProvider
	CharterStore                HarmonicSovereignMeshConstitutionalCharterStore
	MetricsSink                 HarmonicSovereignMeshConstitutionalMetricsSink
}

// -----------------------------------------------------------------------------
// External metric providers
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshConstitutionalGovernanceMetricsProvider interface {
	GetGovernanceMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshConstitutionalTeleologicalMetricsProvider interface {
	GetTeleologicalMetrics(ctx context.Context) (map[string]any, error)
}

// -----------------------------------------------------------------------------
// Constitutional model
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalClause represents a single constitutional rule.
type HarmonicSovereignMeshConstitutionalClause struct {
	ID        HarmonicSovereignMeshConstitutionalClauseID
	Text      string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	Locked    bool
}

// HarmonicSovereignMeshConstitutionalCharter represents the root constitutional document.
type HarmonicSovereignMeshConstitutionalCharter struct {
	ID        HarmonicSovereignMeshConstitutionalCharterID
	Clauses   []*HarmonicSovereignMeshConstitutionalClause
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
	Active    bool
}

// HarmonicSovereignMeshConstitutionalCharterStore persists charters.
type HarmonicSovereignMeshConstitutionalCharterStore interface {
	SaveCharter(ctx context.Context, charter *HarmonicSovereignMeshConstitutionalCharter) error
	LoadCharter(ctx context.Context, id HarmonicSovereignMeshConstitutionalCharterID) (*HarmonicSovereignMeshConstitutionalCharter, error)
	ListCharters(ctx context.Context) ([]*HarmonicSovereignMeshConstitutionalCharter, error)
}

// -----------------------------------------------------------------------------
// Constitutional kernel
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalKernel validates actions against the constitution
// and decides amendments.
type HarmonicSovereignMeshConstitutionalKernel interface {
	EvaluateCharter(
		ctx context.Context,
		charter *HarmonicSovereignMeshConstitutionalCharter,
		govMetrics map[string]any,
		telMetrics map[string]any,
		cfg HarmonicSovereignMeshConstitutionalConfig,
	) (*HarmonicSovereignMeshConstitutionalAmendmentDecision, error)
}

// HarmonicSovereignMeshConstitutionalAmendmentDecision describes charter mutations.
type HarmonicSovereignMeshConstitutionalAmendmentDecision struct {
	CharterID HarmonicSovereignMeshConstitutionalCharterID
	Type      HarmonicSovereignMeshConstitutionalAmendmentType
	Clause    *HarmonicSovereignMeshConstitutionalClause
}

// -----------------------------------------------------------------------------
// Constitutional loop
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalLoop drives continuous constitutional evaluation.
type HarmonicSovereignMeshConstitutionalLoop interface {
	Start(ctx context.Context) error
}

// -----------------------------------------------------------------------------
// Metrics
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalMetrics captures observability for THMS‑CE.
type HarmonicSovereignMeshConstitutionalMetrics struct {
	Timestamp      time.Time
	ActiveCharters int
	AmendedClauses int
	LockedClauses  int
	RemovedClauses int
	ConsensusScore float64
	Tags           map[string]string
}

// HarmonicSovereignMeshConstitutionalMetricsSink receives constitutional metrics.
type HarmonicSovereignMeshConstitutionalMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshConstitutionalMetrics) error
}

// -----------------------------------------------------------------------------
// Constitutional engine (default implementation)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshConstitutionalEngine is the main interface for Phase 116.
type HarmonicSovereignMeshConstitutionalEngine interface {
	Run(ctx context.Context) error
}

// HarmonicSovereignMeshConstitutionalDefaultEngine is the default THMS‑CE implementation.
type HarmonicSovereignMeshConstitutionalDefaultEngine struct {
	cfg    HarmonicSovereignMeshConstitutionalConfig
	deps   HarmonicSovereignMeshConstitutionalDependencies
	kernel HarmonicSovereignMeshConstitutionalKernel
}

// NewHarmonicSovereignMeshConstitutionalDefaultEngine constructs the default engine.
func NewHarmonicSovereignMeshConstitutionalDefaultEngine(
	cfg HarmonicSovereignMeshConstitutionalConfig,
	deps HarmonicSovereignMeshConstitutionalDependencies,
	kernel HarmonicSovereignMeshConstitutionalKernel,
) *HarmonicSovereignMeshConstitutionalDefaultEngine {
	return &HarmonicSovereignMeshConstitutionalDefaultEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
	}
}

// Run executes the constitutional loop.
func (e *HarmonicSovereignMeshConstitutionalDefaultEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			charters, err := e.deps.CharterStore.ListCharters(ctx)
			if err != nil {
				continue
			}

			govMetrics, _ := e.deps.GovernanceMetricsProvider.GetGovernanceMetrics(ctx)
			telMetrics, _ := e.deps.TeleologicalMetricsProvider.GetTeleologicalMetrics(ctx)

			metrics := HarmonicSovereignMeshConstitutionalMetrics{
				Timestamp: time.Now(),
				Tags:      map[string]string{"engine": "HarmonicSovereignMeshConstitutionalDefaultEngine"},
			}

			for _, charter := range charters {
				if !charter.Active {
					continue
				}
				metrics.ActiveCharters++

				decision, err := e.kernel.EvaluateCharter(ctx, charter, govMetrics, telMetrics, e.cfg)
				if err != nil || decision == nil {
					continue
				}

				switch decision.Type {
				case HarmonicSovereignMeshConstitutionalAmendmentTypeAddClause:
					if decision.Clause != nil {
						charter.Clauses = append(charter.Clauses, decision.Clause)
						metrics.AmendedClauses++
					}
				case HarmonicSovereignMeshConstitutionalAmendmentTypeUpdateClause:
					if decision.Clause != nil {
						for i, c := range charter.Clauses {
							if c.ID == decision.Clause.ID {
								charter.Clauses[i] = decision.Clause
								metrics.AmendedClauses++
								break
							}
						}
					}
				case HarmonicSovereignMeshConstitutionalAmendmentTypeRemoveClause:
					if decision.Clause != nil {
						var filtered []*HarmonicSovereignMeshConstitutionalClause
						for _, c := range charter.Clauses {
							if c.ID != decision.Clause.ID {
								filtered = append(filtered, c)
							}
						}
						charter.Clauses = filtered
						metrics.RemovedClauses++
					}

				case HarmonicSovereignMeshConstitutionalAmendmentTypeLockClause:
					if decision.Clause != nil {
						for _, c := range charter.Clauses {
							if c.ID == decision.Clause.ID {
								c.Locked = true
								metrics.LockedClauses++
								break
							}
						}
					}
				}

				charter.UpdatedAt = time.Now()
				_ = e.deps.CharterStore.SaveCharter(ctx, charter)
			}

			if metrics.ActiveCharters == 0 {
				metrics.ConsensusScore = 0
			} else {
				metrics.ConsensusScore =
					float64(metrics.AmendedClauses+metrics.LockedClauses) /
						float64(metrics.ActiveCharters)
			}

			_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
		}
	}
}
