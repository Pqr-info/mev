package main

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------
// Core identifiers & enums
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernancePolicyID uniquely identifies a governance policy.
type HarmonicSovereignMeshAdaptiveGovernancePolicyID string

// HarmonicSovereignMeshAdaptiveGovernanceMutationType describes how a policy mutates.
type HarmonicSovereignMeshAdaptiveGovernanceMutationType string

const (
	HarmonicSovereignMeshAdaptiveGovernanceMutationTypeAdjustThresholds HarmonicSovereignMeshAdaptiveGovernanceMutationType = "adjust_thresholds"
	HarmonicSovereignMeshAdaptiveGovernanceMutationTypeRewriteRules      HarmonicSovereignMeshAdaptiveGovernanceMutationType = "rewrite_rules"
	HarmonicSovereignMeshAdaptiveGovernanceMutationTypeLockPolicy        HarmonicSovereignMeshAdaptiveGovernanceMutationType = "lock_policy"
	HarmonicSovereignMeshAdaptiveGovernanceMutationTypeDeprecatePolicy   HarmonicSovereignMeshAdaptiveGovernanceMutationType = "deprecate_policy"
)

// -----------------------------------------------------------------------------
// Config & dependencies
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernanceConfig configures THMS‑AG behavior.
type HarmonicSovereignMeshAdaptiveGovernanceConfig struct {
	LoopTickInterval          time.Duration
	PolicyMutationSensitivity float64
	ConsensusThreshold        float64
	MaxActivePolicies         int
}

// HarmonicSovereignMeshAdaptiveGovernanceDependencies wires THMS‑AG into Phases 110–113.
type HarmonicSovereignMeshAdaptiveGovernanceDependencies struct {
	SpeciationMetricsProvider HarmonicSovereignMeshAdaptiveGovernanceSpeciationMetricsProvider
	SelectionMetricsProvider  HarmonicSovereignMeshAdaptiveGovernanceSelectionMetricsProvider
	ExtinctionMetricsProvider HarmonicSovereignMeshAdaptiveGovernanceExtinctionMetricsProvider
	PolicyStore               HarmonicSovereignMeshAdaptiveGovernancePolicyStore
	MetricsSink               HarmonicSovereignMeshAdaptiveGovernanceMetricsSink
}

// -----------------------------------------------------------------------------
// External metric providers (Phases 111–113)
// -----------------------------------------------------------------------------

type HarmonicSovereignMeshAdaptiveGovernanceSpeciationMetricsProvider interface {
	GetSpeciationMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshAdaptiveGovernanceSelectionMetricsProvider interface {
	GetSelectionMetrics(ctx context.Context) (map[string]any, error)
}

type HarmonicSovereignMeshAdaptiveGovernanceExtinctionMetricsProvider interface {
	GetExtinctionMetrics(ctx context.Context) (map[string]any, error)
}

// -----------------------------------------------------------------------------
// Governance policy model
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernancePolicy represents a single adaptive policy.
type HarmonicSovereignMeshAdaptiveGovernancePolicy struct {
	ID         HarmonicSovereignMeshAdaptiveGovernancePolicyID
	Rules      map[string]any
	Metadata   map[string]any
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Locked     bool
	Deprecated bool
}

// HarmonicSovereignMeshAdaptiveGovernancePolicyStore persists governance policies.
type HarmonicSovereignMeshAdaptiveGovernancePolicyStore interface {
	SavePolicy(
		ctx context.Context,
		policy *HarmonicSovereignMeshAdaptiveGovernancePolicy,
	) error

	LoadPolicy(
		ctx context.Context,
		id HarmonicSovereignMeshAdaptiveGovernancePolicyID,
	) (*HarmonicSovereignMeshAdaptiveGovernancePolicy, error)

	ListPolicies(
		ctx context.Context,
	) ([]*HarmonicSovereignMeshAdaptiveGovernancePolicy, error)
}

// -----------------------------------------------------------------------------
// Governance kernel
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernanceKernel decides how policies mutate.
type HarmonicSovereignMeshAdaptiveGovernanceKernel interface {
	EvaluatePolicy(
		ctx context.Context,
		policy *HarmonicSovereignMeshAdaptiveGovernancePolicy,
		speciationMetrics map[string]any,
		selectionMetrics map[string]any,
		extinctionMetrics map[string]any,
		cfg HarmonicSovereignMeshAdaptiveGovernanceConfig,
	) (*HarmonicSovereignMeshAdaptiveGovernanceMutationDecision, error)
}

// HarmonicSovereignMeshAdaptiveGovernanceMutationDecision describes policy mutations.
type HarmonicSovereignMeshAdaptiveGovernanceMutationDecision struct {
	PolicyID HarmonicSovereignMeshAdaptiveGovernancePolicyID
	Mutation HarmonicSovereignMeshAdaptiveGovernanceMutationType
	NewRules map[string]any
}

// -----------------------------------------------------------------------------
// Governance loop
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernanceLoop drives continuous governance.
type HarmonicSovereignMeshAdaptiveGovernanceLoop interface {
	Start(ctx context.Context) error
}

// -----------------------------------------------------------------------------
// Metrics
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernanceMetrics captures observability for THMS‑AG.
type HarmonicSovereignMeshAdaptiveGovernanceMetrics struct {
	Timestamp          time.Time
	MutatedPolicies    int
	LockedPolicies     int
	DeprecatedPolicies int
	ConsensusScore     float64
	Tags               map[string]string
}

// HarmonicSovereignMeshAdaptiveGovernanceMetricsSink receives governance metrics.
type HarmonicSovereignMeshAdaptiveGovernanceMetricsSink interface {
	RecordMetrics(ctx context.Context, m HarmonicSovereignMeshAdaptiveGovernanceMetrics) error
}

// -----------------------------------------------------------------------------
// Governance engine (default implementation)
// -----------------------------------------------------------------------------

// HarmonicSovereignMeshAdaptiveGovernanceEngine is the main interface for Phase 114.
type HarmonicSovereignMeshAdaptiveGovernanceEngine interface {
	Run(ctx context.Context) error
}

// HarmonicSovereignMeshAdaptiveGovernanceDefaultEngine is the default THMS‑AG implementation.
type HarmonicSovereignMeshAdaptiveGovernanceDefaultEngine struct {
	cfg    HarmonicSovereignMeshAdaptiveGovernanceConfig
	deps   HarmonicSovereignMeshAdaptiveGovernanceDependencies
	kernel HarmonicSovereignMeshAdaptiveGovernanceKernel
}

// NewHarmonicSovereignMeshAdaptiveGovernanceDefaultEngine constructs the default engine.
func NewHarmonicSovereignMeshAdaptiveGovernanceDefaultEngine(
	cfg HarmonicSovereignMeshAdaptiveGovernanceConfig,
	deps HarmonicSovereignMeshAdaptiveGovernanceDependencies,
	kernel HarmonicSovereignMeshAdaptiveGovernanceKernel,
) *HarmonicSovereignMeshAdaptiveGovernanceDefaultEngine {
	return &HarmonicSovereignMeshAdaptiveGovernanceDefaultEngine{
		cfg:    cfg,
		deps:   deps,
		kernel: kernel,
	}
}

// Run executes the selection loop.
func (e *HarmonicSovereignMeshAdaptiveGovernanceDefaultEngine) Run(ctx context.Context) error {
	ticker := time.NewTicker(e.cfg.LoopTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			speciationMetrics, _ := e.deps.SpeciationMetricsProvider.GetSpeciationMetrics(ctx)
			selectionMetrics, _ := e.deps.SelectionMetricsProvider.GetSelectionMetrics(ctx)
			extinctionMetrics, _ := e.deps.ExtinctionMetricsProvider.GetExtinctionMetrics(ctx)

			policies, err := e.deps.PolicyStore.ListPolicies(ctx)
			if err != nil {
				continue
			}

			metrics := HarmonicSovereignMeshAdaptiveGovernanceMetrics{
				Timestamp: time.Now(),
				Tags:      map[string]string{"engine": "HarmonicSovereignMeshAdaptiveGovernanceDefaultEngine"},
			}

			for _, policy := range policies {
				decision, err := e.kernel.EvaluatePolicy(ctx, policy, speciationMetrics, selectionMetrics, extinctionMetrics, e.cfg)
				if err != nil {
					continue
				}

				switch decision.Mutation {
				case HarmonicSovereignMeshAdaptiveGovernanceMutationTypeAdjustThresholds, HarmonicSovereignMeshAdaptiveGovernanceMutationTypeRewriteRules:
					policy.Rules = decision.NewRules
					policy.UpdatedAt = time.Now()
					metrics.MutatedPolicies++
				case HarmonicSovereignMeshAdaptiveGovernanceMutationTypeLockPolicy:
					policy.Locked = true
					policy.UpdatedAt = time.Now()
					metrics.LockedPolicies++
				case HarmonicSovereignMeshAdaptiveGovernanceMutationTypeDeprecatePolicy:
					policy.Deprecated = true
					policy.UpdatedAt = time.Now()
					metrics.DeprecatedPolicies++
				}

				_ = e.deps.PolicyStore.SavePolicy(ctx, policy)
			}

			metrics.ConsensusScore = float64(metrics.MutatedPolicies) / float64(len(policies)+1)
			_ = e.deps.MetricsSink.RecordMetrics(ctx, metrics)
		}
	}
}
