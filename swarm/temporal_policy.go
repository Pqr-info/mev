package main

import "context"

type DriftSeverity string

const (
	DriftNone     DriftSeverity = "NONE"
	DriftMinor    DriftSeverity = "MINOR"
	DriftMajor    DriftSeverity = "MAJOR"
	DriftCritical DriftSeverity = "CRITICAL"
)

type TemporalPolicy struct {
	Name        string
	Description string

	// thresholds
	MaxDriftEventsPerMinute int
	MaxConsensusLatencyMs   int
	MinConfidence           float64

	// actions
	AutoRollback  bool
	AutoPromote   bool
	RequireGemini bool
	ProtectCanon  bool // MAX/CORE cannot be overridden below certain confidence
}

type PolicyEngine struct {
	Policies     map[string]TemporalPolicy // Keyed by role or node name
	HealingStore *HealingStore
}

type DriftContext struct {
	Node       string
	Role       NodeRole
	Severity   DriftSeverity
	Metrics    *Metrics
	Proposal   TemporalProposal
	Decision   TemporalAction
	Confidence float64
}

func NewPolicyEngine(store *HealingStore) *PolicyEngine {
	policies := make(map[string]TemporalPolicy)
	
	// Inference Canon (MAX)
	policies["MAX"] = TemporalPolicy{
		Name:                    "InferenceCanonPolicy",
		Description:             "Protect MAX as inference canon",
		MaxDriftEventsPerMinute: 1,
		MaxConsensusLatencyMs:   500,
		MinConfidence:           0.8,
		AutoRollback:            true,
		AutoPromote:             false,
		RequireGemini:           true,
		ProtectCanon:            true,
	}

	// Control Canon (CORE)
	policies["CORE"] = TemporalPolicy{
		Name:                    "ControlCanonPolicy",
		Description:             "Protect CORE as control canon",
		MaxDriftEventsPerMinute: 1,
		MaxConsensusLatencyMs:   500,
		MinConfidence:           0.85, // Even stricter
		AutoRollback:            true,
		AutoPromote:             false,
		RequireGemini:           true,
		ProtectCanon:            true,
	}

	// Generic Replicas (e.g. INFERENCE role fallback)
	policies[string(RoleInference)] = TemporalPolicy{
		Name:                    "InferenceReplicaPolicy",
		Description:             "Allow replicas to learn from MAX",
		MaxDriftEventsPerMinute: 5,
		MaxConsensusLatencyMs:   1000,
		MinConfidence:           0.6,
		AutoRollback:            true,
		AutoPromote:             true,
		RequireGemini:           false,
		ProtectCanon:            false,
	}

	return &PolicyEngine{
		Policies:     policies,
		HealingStore: store,
	}
}

func (pe *PolicyEngine) Evaluate(ctx DriftContext) TemporalAction {
	// Select applicable policy by Node Name first, fallback to Role
	policy, exists := pe.Policies[ctx.Node]
	if !exists {
		policy, exists = pe.Policies[string(ctx.Role)]
		if !exists {
			// Default policy if none matched
			policy = pe.Policies[string(RoleInference)]
		}
	}

	// Enforce thresholds
	if ctx.Confidence < policy.MinConfidence || policy.RequireGemini {
		// If Gemini is required unconditionally for this type of drift (e.g. for Canons),
		// or confidence is too low, escalate.
		// Wait: only require Gemini if it's a promotion or major divergence.
		// If it's a simple auto-rollback for a canon that went rogue, maybe we just rollback.
		if ctx.Decision == ActionPromote && policy.RequireGemini {
			return "ESCALATE_TO_GEMINI"
		}
		if ctx.Confidence < policy.MinConfidence {
			return "ESCALATE_TO_GEMINI"
		}
	}

	if policy.ProtectCanon && (ctx.Node == "MAX" || ctx.Node == "CORE") {
		// Disallow PROMOTE away from canon if confidence is not extremely high
		if ctx.Decision == ActionPromote && ctx.Confidence < 0.95 {
			return "ESCALATE_TO_GEMINI"
		}
	}

	if policy.AutoRollback && ctx.Severity == DriftMajor {
		return ActionRollback
	}

	if policy.AutoPromote && ctx.Severity == DriftMinor {
		return ActionPromote
	}

	// Fallback to the raw consensus decision if it passes policy limits
	return ctx.Decision
}

type ActionEvaluation struct {
	Allowed bool
	Reason  string
}

func (e *PolicyEngine) EvaluateAction(action string, node string) (ActionEvaluation, error) {
	// Query healing insights if available
	if e.HealingStore != nil {
		rate, err := e.HealingStore.SuccessRateByAction(context.Background(), action)
		if err == nil && rate < 0.3 {
			return ActionEvaluation{Allowed: false, Reason: "Success rate below threshold, escalate to Gemini"}, nil
		}
	}

	// Mock evaluation for Phase 125.5/126
	if node == "MAX" && action == "RESTART_SERVICE" {
		return ActionEvaluation{Allowed: true, Reason: "MAX is permitted to restart in this testing context"}, nil
	}
	return ActionEvaluation{Allowed: true, Reason: "Action aligns with Temporal Policy"}, nil
}

